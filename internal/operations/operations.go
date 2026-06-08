package operations

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/platform/pkg/contracts"
)

type AuditEvent struct {
	Action    string         `json:"action"`
	Actor     string         `json:"actor"`
	Resource  string         `json:"resource"`
	Details   map[string]any `json:"details,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type ErrorRecord struct {
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu     sync.Mutex
	audit  []AuditEvent
	errors []ErrorRecord
	limit  int
}

func NewStore(limit int) *Store {
	if limit <= 0 {
		limit = 50
	}
	return &Store{limit: limit}
}

func (s *Store) RecordAudit(event AuditEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event.CreatedAt = time.Now().UTC()
	event.Details = redact(event.Details)
	s.audit = appendBounded(s.audit, event, s.limit)
}

func (s *Store) Audit() []AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AuditEvent{}, s.audit...)
}

func (s *Store) RecordError(record ErrorRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record.CreatedAt = time.Now().UTC()
	s.errors = appendBounded(s.errors, record, s.limit)
}

func (s *Store) Errors() []ErrorRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ErrorRecord{}, s.errors...)
}

type HealthResult struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func Health(ctx context.Context, provider storage.StoreProvider, ops *Store, runtimeState string) []HealthResult {
	results := []HealthResult{
		{ID: "plugin_runtime", Status: "ok", Message: runtimeState},
		{ID: "audit", Status: "ok"},
		{ID: "diagnostics", Status: "ok"},
	}
	err := provider.ForWorkspace("root").WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		_, err := repos.ContentTypes.List(ctx)
		return err
	})
	if err != nil {
		results = append(results, HealthResult{ID: "storage", Status: "error", Message: err.Error()})
		if ops != nil {
			ops.RecordError(ErrorRecord{Source: "health.storage", Message: err.Error()})
		}
	} else {
		results = append(results, HealthResult{ID: "storage", Status: "ok"})
	}
	return results
}

type Snapshot struct {
	Version      string             `json:"version"`
	ExportedAt   time.Time          `json:"exported_at"`
	Posts        []contracts.Record `json:"posts"`
	Pages        []contracts.Record `json:"pages"`
	ContentTypes []contracts.Record `json:"content_types"`
	Taxonomies   []contracts.Record `json:"taxonomies"`
	Terms        []contracts.Record `json:"terms"`
	Media        []contracts.Record `json:"media"`
	Authors      []contracts.Record `json:"authors"`
	Settings     []contracts.Record `json:"settings"`
	Menus        []contracts.Record `json:"menus"`
}

const SnapshotVersion = "gocms.snapshot.v1"

func ExportSnapshot(ctx context.Context, provider storage.StoreProvider) (Snapshot, error) {
	snapshot := Snapshot{Version: SnapshotVersion, ExportedAt: time.Now().UTC()}
	err := provider.ForWorkspace("root").WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		var err error
		if snapshot.ContentTypes, err = list(ctx, repos.ContentTypes); err != nil {
			return err
		}
		if snapshot.Taxonomies, err = list(ctx, repos.Taxonomies); err != nil {
			return err
		}
		if snapshot.Terms, err = list(ctx, repos.Terms); err != nil {
			return err
		}
		if snapshot.Authors, err = list(ctx, repos.Authors); err != nil {
			return err
		}
		if snapshot.Posts, err = list(ctx, repos.Posts); err != nil {
			return err
		}
		if snapshot.Pages, err = list(ctx, repos.Pages); err != nil {
			return err
		}
		if snapshot.Media, err = list(ctx, repos.MediaAssets); err != nil {
			return err
		}
		if snapshot.Settings, err = list(ctx, repos.Settings); err != nil {
			return err
		}
		snapshot.Menus, err = list(ctx, repos.Menus)
		return err
	})
	return snapshot, err
}

func ImportSnapshot(ctx context.Context, provider storage.StoreProvider, snapshot Snapshot) error {
	return provider.ForWorkspace("root").WithinTx(ctx, func(ctx context.Context, repos storage.Repositories) error {
		for _, group := range []struct {
			repo  storage.RecordRepository
			items []contracts.Record
		}{
			{repos.ContentTypes, snapshot.ContentTypes},
			{repos.Taxonomies, snapshot.Taxonomies},
			{repos.Terms, snapshot.Terms},
			{repos.Authors, snapshot.Authors},
			{repos.Posts, snapshot.Posts},
			{repos.Pages, snapshot.Pages},
			{repos.MediaAssets, snapshot.Media},
			{repos.Settings, snapshot.Settings},
			{repos.Menus, snapshot.Menus},
		} {
			for _, item := range group.items {
				id, _ := item["id"].(string)
				if id == "" {
					id, _ = item["key"].(string)
				}
				if id == "" {
					id, _ = item["location"].(string)
				}
				if id == "" {
					continue
				}
				if err := group.repo.Put(ctx, contracts.RecordID(id), item); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func list(ctx context.Context, repo storage.RecordRepository) ([]contracts.Record, error) {
	page, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return page.Records, nil
}

func redact(details map[string]any) map[string]any {
	if details == nil {
		return nil
	}
	redacted := map[string]any{}
	for key, value := range details {
		if strings.Contains(strings.ToLower(key), "token") || strings.Contains(strings.ToLower(key), "secret") || strings.Contains(strings.ToLower(key), "password") {
			redacted[key] = "[redacted]"
			continue
		}
		redacted[key] = value
	}
	return redacted
}

func appendBounded[T any](items []T, item T, limit int) []T {
	items = append(items, item)
	if len(items) > limit {
		return items[len(items)-limit:]
	}
	return items
}
