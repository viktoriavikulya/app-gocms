package taxonomy_test

import (
	"context"
	"testing"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	apptaxonomy "github.com/fastygo/app-gocms/internal/application/taxonomy"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts/contractstest"
)

func TestTaxonomyAssignTerms(t *testing.T) {
	provider := storage.NewProvider(contractstest.NewMemoryStorage())
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos storage.Repositories) error {
		appRepos := storage.NewApplicationRepositories(repos)
		if err := appcontenttype.NewService(appRepos).InstallBuiltIns(ctx); err != nil {
			return err
		}
		if _, err := appcontent.NewService(appRepos, appRepos).CreateDraft(ctx, content.Entry{ID: "post-1", Kind: content.KindPost, Title: map[string]string{"en": "Hello"}, Slug: "hello"}); err != nil {
			return err
		}
		service := apptaxonomy.NewService(appRepos, appRepos)
		if err := service.Register(ctx, taxonomy.Definition{Type: "category", Label: "Category", Mode: taxonomy.ModeHierarchical, Public: true}); err != nil {
			return err
		}
		if err := service.CreateTerm(ctx, taxonomy.Term{ID: "news", TaxonomyType: "category", Name: "News", Slug: "news"}); err != nil {
			return err
		}
		entry, err := service.AssignTerms(ctx, "post-1", []string{"news"})
		if err != nil {
			return err
		}
		if len(entry.TermIDs) != 1 || entry.TermIDs[0] != "news" {
			t.Fatalf("terms = %#v", entry.TermIDs)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
