package sqlite

import (
	"context"
	"time"

	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/contracts"
)

func SeedMinimalSite(ctx context.Context, store *Store, workspace contracts.WorkspaceID) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
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
			{string(records.RecordPost), "hello-world", contracts.Record{"title": map[string]string{"en": "Hello world"}, "slug": "hello-world", "status": "published", "visibility": "public", "author_id": "admin", "created_at": now, "updated_at": now}},
			{string(records.RecordPage), "about", contracts.Record{"title": map[string]string{"en": "About"}, "slug": "about", "status": "published", "visibility": "public", "author_id": "admin", "created_at": now, "updated_at": now}},
			{string(RecordSetting), "site", contracts.Record{"key": "site", "title": "AppCMS", "locale": "en"}},
			{string(RecordMenu), "primary", contracts.Record{"location": "primary", "items": []string{"about"}}},
		}
		for _, item := range seed {
			if err := tx.Put(ctx, item.recordType, item.id, item.record); err != nil {
				return err
			}
		}
		return nil
	})
}
