package appschema_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	"github.com/fastygo/app-gocms/internal/appschema"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestHydratorListViewerHidesDrafts(t *testing.T) {
	registry := appschema.MustRegistry()
	screen, err := registry.Screen("/go-admin/posts")
	if err != nil {
		t.Fatal(err)
	}
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	past := time.Now().UTC().Add(-2 * time.Hour)
	err = provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		service := appcontent.NewService(appRepos, appRepos).WithClock(func() time.Time { return past })
		if _, err := service.CreateDraft(ctx, content.Entry{ID: "draft-post", Kind: content.KindPost, Title: map[string]string{"en": "Draft"}, Slug: "draft-post"}); err != nil {
			return err
		}
		published, err := service.CreateDraft(ctx, content.Entry{ID: "pub-post", Kind: content.KindPost, Title: map[string]string{"en": "Published"}, Slug: "pub-post"})
		if err != nil {
			return err
		}
		if _, err := service.Publish(ctx, published.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	viewer := principalForRole(appauthn.RoleViewer)
	hydrator := appschema.Hydrator{Provider: provider, Workspace: "root"}
	hydrated, err := hydrator.Hydrate(context.Background(), screen, "/go-admin/posts", url.Values{}, viewer)
	if err != nil {
		t.Fatal(err)
	}
	if len(hydrated.Rows) != 1 || hydrated.Rows[0].ID != "pub-post" {
		t.Fatalf("viewer rows = %#v, want only published post", hydrated.Rows)
	}

	admin := principalForRole(appauthn.RoleAdmin)
	hydrated, err = hydrator.Hydrate(context.Background(), screen, "/go-admin/posts", url.Values{}, admin)
	if err != nil {
		t.Fatal(err)
	}
	if len(hydrated.Rows) != 2 {
		t.Fatalf("admin rows = %d, want 2", len(hydrated.Rows))
	}
	if hydrated.Pagination == nil || hydrated.Pagination.Total != 2 || hydrated.Pagination.TotalPages != 1 {
		t.Fatalf("pagination = %#v", hydrated.Pagination)
	}
}

func TestHydratorEditPrefillsForm(t *testing.T) {
	registry := appschema.MustRegistry()
	screen, err := registry.Screen("/go-admin/posts/post-1/edit")
	if err != nil {
		t.Fatal(err)
	}
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err = provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		service := appcontent.NewService(appRepos, appRepos)
		_, err := service.CreateDraft(ctx, content.Entry{
			ID: "post-1", Kind: content.KindPost,
			Title: map[string]string{"en": "Hello"}, Slug: "hello",
			Content: "Body", AuthorID: "admin",
		})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	admin := principalForRole(appauthn.RoleAdmin)
	hydrator := appschema.Hydrator{Provider: provider, Workspace: "root"}
	hydrated, err := hydrator.Hydrate(context.Background(), screen, "/go-admin/posts/post-1/edit", url.Values{}, admin)
	if err != nil {
		t.Fatal(err)
	}
	if hydrated.FormRecord == nil || hydrated.FormRecord.Values["title"] != "Hello" || hydrated.FormRecord.Values["content"] != "Body" {
		t.Fatalf("form values = %#v", hydrated.FormRecord)
	}
	if len(hydrated.RowActions) == 0 {
		t.Fatalf("expected workflow actions on edit screen")
	}
}

func principalForRole(role string) appauthn.Principal {
	store, err := appauthn.NewSeededMemoryStore()
	if err != nil {
		panic(err)
	}
	service := appauthn.NewService(store)
	id := role
	if role == appauthn.RoleAdmin {
		id = "admin"
	}
	if role == appauthn.RoleViewer {
		id = "viewer"
	}
	principal, ok := service.Principal(contracts.PrincipalID(id))
	if !ok {
		panic("missing seeded principal " + role)
	}
	return principal
}
