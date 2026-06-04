package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
)

type migration struct {
	Version     string
	Description string
	Statements  []string
}

func sqliteMigrations() []migration {
	return []migration{
		{
			Version:     "0001_cms_schema",
			Description: "create workspace-scoped CMS records tables",
			Statements: append(recordTableStatements(),
				`CREATE INDEX IF NOT EXISTS idx_posts_workspace_status ON posts(workspace_id, status)`,
				`CREATE INDEX IF NOT EXISTS idx_pages_workspace_status ON pages(workspace_id, status)`,
				`CREATE INDEX IF NOT EXISTS idx_terms_workspace_slug ON terms(workspace_id, slug)`,
				`CREATE INDEX IF NOT EXISTS idx_media_assets_workspace_slug ON media_assets(workspace_id, slug)`,
				`CREATE INDEX IF NOT EXISTS idx_authors_workspace_slug ON authors(workspace_id, slug)`,
			),
		},
	}
}

func recordTableStatements() []string {
	statements := []string{}
	for _, table := range cmsTables() {
		statements = append(statements, fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				workspace_id TEXT NOT NULL,
				id TEXT NOT NULL,
				record_type TEXT NOT NULL,
				status TEXT NOT NULL DEFAULT '',
				slug TEXT NOT NULL DEFAULT '',
				updated_at TEXT NOT NULL DEFAULT '',
				payload_json TEXT NOT NULL,
				PRIMARY KEY(workspace_id, id)
			)
		`, table.Name))
		statements = append(statements, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_workspace_type ON %s(workspace_id, record_type)`, table.Name, table.Name))
	}
	return statements
}

func (s *Store) applyMigrations(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, description TEXT NOT NULL, checksum TEXT NOT NULL, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return err
	}
	for _, migration := range sqliteMigrations() {
		if slices.Contains(applied, migration.Version) {
			continue
		}
		if err := s.applyMigration(ctx, migration); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) appliedMigrations(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []string{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		result = append(result, version)
	}
	return result, rows.Err()
}

func (s *Store) applyMigration(ctx context.Context, migration migration) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, statement := range migration.Statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Version, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, description, checksum, applied_at) VALUES (?, ?, ?, ?)`, migration.Version, migration.Description, migrationChecksum(migration), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Version, err)
	}
	return tx.Commit()
}

func (s *Store) MigrationStatus(ctx context.Context) error {
	applied, err := s.appliedMigrations(ctx)
	if err != nil {
		return err
	}
	for _, migration := range sqliteMigrations() {
		if !slices.Contains(applied, migration.Version) {
			return fmt.Errorf("pending migration %s", migration.Version)
		}
	}
	return nil
}

func migrationChecksum(migration migration) string {
	sum := sha256.Sum256([]byte(strings.Join(migration.Statements, "\n")))
	return hex.EncodeToString(sum[:])
}
