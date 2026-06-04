package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	appstorage "github.com/fastygo/app-gocms/internal/storage"
	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/contracts"
)

func TestInitAppliesMigrationsAndIsIdempotent(t *testing.T) {
	store := openTestStore(t)
	defer closeTestStore(t, store)

	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := store.MigrationStatus(context.Background()); err != nil {
		t.Fatal(err)
	}

	var count int
	err := store.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM schema_migrations`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != len(sqliteMigrations()) {
		t.Fatalf("migration count = %d, want %d", count, len(sqliteMigrations()))
	}
}

func TestInitCreatesCMSTables(t *testing.T) {
	store := openTestStore(t)
	defer closeTestStore(t, store)

	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, table := range cmsTables() {
		var name string
		err := store.db.QueryRowContext(context.Background(), `SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table.Name).Scan(&name)
		if err == sql.ErrNoRows {
			t.Fatalf("table %s was not created", table.Name)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestStoragePortCRUDAndWorkspaceIsolation(t *testing.T) {
	store := openInitializedTestStore(t)
	defer closeTestStore(t, store)

	err := store.WithinWorkspaceTx(context.Background(), "root", func(ctx context.Context, tx contracts.StorageTx) error {
		return tx.Put(ctx, string(records.RecordPost), "post-1", contracts.Record{"title": "Root post", "status": "published"})
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.WithinWorkspaceTx(context.Background(), "sales", func(ctx context.Context, tx contracts.StorageTx) error {
		page, err := tx.List(ctx, contracts.Query{Workspace: "sales", RecordType: string(records.RecordPost)})
		if err != nil {
			return err
		}
		if page.TotalItems != 0 {
			t.Fatalf("sales workspace sees %d root posts, want 0", page.TotalItems)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = store.WithinWorkspaceTx(context.Background(), "root", func(ctx context.Context, tx contracts.StorageTx) error {
		got, err := tx.Get(ctx, string(records.RecordPost), "post-1")
		if err != nil {
			return err
		}
		if got["id"] != "post-1" || got["status"] != "published" {
			t.Fatalf("record = %#v, want id/status", got)
		}
		if err := tx.Delete(ctx, string(records.RecordPost), "post-1"); err != nil {
			return err
		}
		page, err := tx.List(ctx, contracts.Query{Workspace: "root", RecordType: string(records.RecordPost)})
		if err != nil {
			return err
		}
		if page.TotalItems != 0 {
			t.Fatalf("root workspace sees %d posts after delete, want 0", page.TotalItems)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStoragePortRejectsUnsupportedRecordType(t *testing.T) {
	store := openInitializedTestStore(t)
	defer closeTestStore(t, store)

	err := store.WithinWorkspaceTx(context.Background(), "root", func(ctx context.Context, tx contracts.StorageTx) error {
		return tx.Put(ctx, "plugin_record", "record-1", contracts.Record{})
	})
	if err == nil {
		t.Fatal("unsupported record type should fail")
	}
}

func TestSeedMinimalSiteAndRepositoryWrapper(t *testing.T) {
	store := openInitializedTestStore(t)
	defer closeTestStore(t, store)

	if err := SeedMinimalSite(context.Background(), store, "root"); err != nil {
		t.Fatal(err)
	}

	provider := appstorage.NewProvider(store)
	err := provider.ForWorkspace("root").WithinTx(context.Background(), func(ctx context.Context, repos appstorage.Repositories) error {
		posts, err := repos.Posts.List(ctx)
		if err != nil {
			return err
		}
		pages, err := repos.Pages.List(ctx)
		if err != nil {
			return err
		}
		settings, err := repos.Settings.List(ctx)
		if err != nil {
			return err
		}
		if posts.TotalItems != 1 || pages.TotalItems != 1 || settings.TotalItems != 1 {
			t.Fatalf("seed counts posts=%d pages=%d settings=%d, want 1/1/1", posts.TotalItems, pages.TotalItems, settings.TotalItems)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func openInitializedTestStore(t *testing.T) *Store {
	t.Helper()
	store := openTestStore(t)
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	return store
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "appcms.db"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func closeTestStore(t *testing.T, store *Store) {
	t.Helper()
	if err := store.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
