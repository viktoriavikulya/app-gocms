package rest

import (
	"testing"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/app-gocms/internal/domain/users"
)

func TestContentProjectionStripsPrivateMetadata(t *testing.T) {
	entry := domaincontent.Entry{
		ID:     "post-1",
		Kind:   domaincontent.KindPost,
		Status: domaincontent.StatusPublished,
		Slug:   "hello",
		Title:  map[string]string{"en": "Hello"},
		Metadata: map[string]any{
			"seo_title":    "Public",
			"note_private": "secret",
		},
	}
	public := ContentProjection(entry, false)
	if _, ok := public.Metadata["note_private"]; ok {
		t.Fatalf("private metadata leaked: %#v", public.Metadata)
	}
	if public.Metadata["seo_title"] != "Public" {
		t.Fatalf("public metadata missing: %#v", public.Metadata)
	}
	private := ContentProjection(entry, true)
	if private.Metadata["note_private"] != "secret" {
		t.Fatalf("private metadata missing when allowed: %#v", private.Metadata)
	}
}

func TestContentProjectionTaxonomyIDsAndLinks(t *testing.T) {
	entry := domaincontent.Entry{
		ID:      "post-1",
		Kind:    domaincontent.KindPost,
		Status:  domaincontent.StatusPublished,
		Slug:    "hello",
		TermIDs: []string{"news", "featured"},
	}
	dto := ContentProjection(entry, false)
	if len(dto.TaxonomyIDs) != 2 || dto.TaxonomyIDs[0] != "news" {
		t.Fatalf("taxonomy_ids = %#v", dto.TaxonomyIDs)
	}
	if dto.Links["self"] != "/go-json/go/v2/posts/post-1" {
		t.Fatalf("links.self = %q", dto.Links["self"])
	}
}

func TestLocalizedMaps(t *testing.T) {
	if got := localize("title"); got["en"] != "title" {
		t.Fatalf("localize = %#v", got)
	}
	if got := delocalize(map[string]string{"en": "title"}); got != "title" {
		t.Fatalf("delocalize = %q", got)
	}
	entry := domaincontent.Entry{Title: map[string]string{"en": "Hello", "ru": "Привет"}}
	dto := ContentProjection(entry, false)
	if dto.Title["en"] != "Hello" || dto.Title["ru"] != "Привет" {
		t.Fatalf("title map = %#v", dto.Title)
	}
}

func TestProjectPublicSettingsStripsPrivateKeys(t *testing.T) {
	items := []settings.Value{
		{Key: "site.title", Value: "AppCMS", Public: true},
		{Key: "site.secret_private", Value: "hidden", Public: false, Private: true},
		{Key: "site.note", Value: "ok", Public: true},
	}
	projected := projectPublicSettings(items)
	if len(projected) != 2 {
		t.Fatalf("projected settings = %#v", projected)
	}
	for _, item := range projected {
		if stringsHasSuffix(item.Key, "_private") {
			t.Fatalf("private key leaked: %q", item.Key)
		}
	}
}

func stringsHasSuffix(value, suffix string) bool {
	return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
}

func TestMenuProjectionBuildsTree(t *testing.T) {
	menu := menus.Menu{
		ID:       "main",
		Location: "primary",
		Items: []menus.Item{
			{ID: "a", Label: "Home", URL: "/"},
			{ID: "b", Label: "About", URL: "/about", ParentID: "a"},
			{ID: "c", Label: "Blog", URL: "/blog"},
		},
	}
	dto := MenuProjection(menu)
	if len(dto.Items) != 2 {
		t.Fatalf("root items = %d", len(dto.Items))
	}
	if len(dto.Items[0].Children) != 1 || dto.Items[0].Children[0].Label != "About" {
		t.Fatalf("nested menu = %#v", dto.Items)
	}
}

func TestIsPublicEntry(t *testing.T) {
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	public := domaincontent.Entry{
		Status:     domaincontent.StatusPublished,
		Visibility: domaincontent.VisibilityPublic,
		PublishedAt: now.Add(-time.Hour),
	}
	if !IsPublicEntry(public, now) {
		t.Fatal("expected published public entry to be public")
	}
	draft := public
	draft.Status = domaincontent.StatusDraft
	if IsPublicEntry(draft, now) {
		t.Fatal("draft should not be public")
	}
}

func TestTaxonomyAndMediaProjections(t *testing.T) {
	tax := TaxonomyProjection(taxonomy.Definition{Type: "category", Label: "Category", Mode: taxonomy.ModeFlat, AssignedKinds: []string{"post"}, Public: true})
	if tax.Type != "category" || len(tax.AssignedToKinds) != 1 {
		t.Fatalf("taxonomy dto = %#v", tax)
	}
	term := TermProjection(taxonomy.Term{ID: "news", TaxonomyType: "category", Name: "News", Slug: "news"})
	if term.Name["en"] != "News" {
		t.Fatalf("term name = %#v", term.Name)
	}
	mediaDTO := MediaProjection(media.Asset{ID: "m1", Title: "hero.png", MIMEType: "image/png"})
	if mediaDTO.Filename != "hero.png" {
		t.Fatalf("media dto = %#v", mediaDTO)
	}
	author := AuthorProjection(domainusers.AuthorProfile{ID: "admin", Slug: "admin", DisplayName: "Admin"}, "/img/admin.png")
	if author.AvatarURL != "/img/admin.png" {
		t.Fatalf("author dto = %#v", author)
	}
	typeDTO := ContentTypeProjection(contenttype.BuiltInPost())
	if typeDTO.ID != "post" {
		t.Fatalf("content type dto = %#v", typeDTO)
	}
}
