package sqlite

import (
	"context"
	"time"

	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/contracts"
)

var seedNow = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)

func SeedMinimalSite(ctx context.Context, store *Store, workspace contracts.WorkspaceID) error {
	return store.WithinWorkspaceTx(ctx, workspace, func(ctx context.Context, tx contracts.StorageTx) error {
		seed := []struct {
			recordType string
			id         contracts.RecordID
			record     contracts.Record
		}{
			{string(records.RecordContentType), "post", contracts.Record{"label": "Post", "permalink_pattern": "/posts/{slug}", "public": true}},
			{string(records.RecordContentType), "page", contracts.Record{"label": "Page", "permalink_pattern": "/{slug}", "public": true}},
			{string(records.RecordTaxonomy), "category", contracts.Record{"type": "category", "label": "Category", "mode": "hierarchical", "public": true}},
			{string(records.RecordTerm), "news", contracts.Record{"taxonomy_type": "category", "name": "News", "slug": "news"}},
			{string(records.RecordAuthor), "admin", contracts.Record{"display_name": "Admin", "slug": "admin", "active": true}},
			{string(records.RecordPost), "hello-world", contracts.Record{"title": map[string]string{"en": "Hello world"}, "slug": "hello-world", "status": "published", "visibility": "public", "author_id": "admin", "created_at": seedNow, "updated_at": seedNow}},
			{string(records.RecordPost), "post-rest", contracts.Record{"title": map[string]string{"en": "REST Post"}, "slug": "rest-post", "content": "Hello REST", "status": "published", "visibility": "public", "author_id": "admin", "created_at": seedNow, "updated_at": seedNow}},
			{string(records.RecordPage), "about", contracts.Record{"title": map[string]string{"en": "About"}, "slug": "about", "status": "published", "visibility": "public", "author_id": "admin", "created_at": seedNow, "updated_at": seedNow}},
			{string(RecordSetting), "site.title", contracts.Record{"key": "site.title", "group": "site", "value": "AppCMS", "public": true}},
			{string(RecordSetting), "site.description", contracts.Record{"key": "site.description", "group": "site", "value": "A FastyGo powered site.", "public": true}},
			{string(RecordSetting), "theme.active", contracts.Record{"key": "theme.active", "group": "theme", "value": "gocms-default", "public": true}},
			{string(RecordMenu), "primary", contracts.Record{"id": "primary", "location": "primary", "items": []map[string]any{{"id": "about", "label": "About", "url": "/about"}}}},
		}
		for _, item := range seed {
			if err := tx.Put(ctx, item.recordType, item.id, item.record); err != nil {
				return err
			}
		}
		return nil
	})
}
