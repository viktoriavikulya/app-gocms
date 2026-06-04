package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fastygo/platform/pkg/contracts"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(dataSource string) (*Store, error) {
	if strings.TrimSpace(dataSource) == "" || dataSource == "fixture" {
		dataSource = "file:appcms.db"
	}
	db, err := sql.Open("sqlite", dataSource)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func (s *Store) Init(ctx context.Context) error {
	return s.applyMigrations(ctx)
}

func (s *Store) Close(context.Context) error {
	return s.db.Close()
}

func (s *Store) WithinWorkspaceTx(ctx context.Context, workspace contracts.WorkspaceID, fn func(context.Context, contracts.StorageTx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(ctx, storageTx{workspace: workspace, tx: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

type storageTx struct {
	workspace contracts.WorkspaceID
	tx        *sql.Tx
}

func (t storageTx) List(ctx context.Context, query contracts.Query) (contracts.PageResult, error) {
	if query.Workspace != "" && query.Workspace != t.workspace {
		return contracts.PageResult{}, fmt.Errorf("query workspace %q does not match tx workspace %q", query.Workspace, t.workspace)
	}
	table, err := tableForRecordType(query.RecordType)
	if err != nil {
		return contracts.PageResult{}, err
	}
	rows, err := t.tx.QueryContext(ctx, fmt.Sprintf(`SELECT payload_json FROM %s WHERE workspace_id = ? ORDER BY id`, table.Name), string(t.workspace))
	if err != nil {
		return contracts.PageResult{}, err
	}
	defer rows.Close()

	records := []contracts.Record{}
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return contracts.PageResult{}, err
		}
		record, err := decodeRecord(payload)
		if err != nil {
			return contracts.PageResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return contracts.PageResult{}, err
	}

	page, perPage := normalizePagination(query.Page, query.PerPage)
	start := (page - 1) * perPage
	total := len(records)
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return contracts.PageResult{Records: records[start:end], Page: page, PerPage: perPage, TotalItems: total}, nil
}

func (t storageTx) Get(ctx context.Context, recordType string, id contracts.RecordID) (contracts.Record, error) {
	table, err := tableForRecordType(recordType)
	if err != nil {
		return nil, err
	}
	var payload string
	err = t.tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT payload_json FROM %s WHERE workspace_id = ? AND id = ?`, table.Name), string(t.workspace), string(id)).Scan(&payload)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%s record %q not found in workspace %q", recordType, id, t.workspace)
	}
	if err != nil {
		return nil, err
	}
	return decodeRecord(payload)
}

func (t storageTx) Put(ctx context.Context, recordType string, id contracts.RecordID, record contracts.Record) error {
	table, err := tableForRecordType(recordType)
	if err != nil {
		return err
	}
	payloadRecord := cloneRecord(record)
	payloadRecord["id"] = string(id)
	payload, err := json.Marshal(payloadRecord)
	if err != nil {
		return err
	}
	_, err = t.tx.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO %s (workspace_id, id, record_type, status, slug, updated_at, payload_json)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(workspace_id, id) DO UPDATE SET
			record_type = excluded.record_type,
			status = excluded.status,
			slug = excluded.slug,
			updated_at = excluded.updated_at,
			payload_json = excluded.payload_json
	`, table.Name), string(t.workspace), string(id), recordType, stringField(payloadRecord, "status"), stringField(payloadRecord, "slug"), stringField(payloadRecord, "updated_at"), string(payload))
	return err
}

func (t storageTx) Delete(ctx context.Context, recordType string, id contracts.RecordID) error {
	table, err := tableForRecordType(recordType)
	if err != nil {
		return err
	}
	_, err = t.tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM %s WHERE workspace_id = ? AND id = ?`, table.Name), string(t.workspace), string(id))
	return err
}

func decodeRecord(payload string) (contracts.Record, error) {
	record := contracts.Record{}
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return nil, err
	}
	return record, nil
}

func cloneRecord(record contracts.Record) contracts.Record {
	clone := contracts.Record{}
	for key, value := range record {
		clone[key] = value
	}
	return clone
}

func stringField(record contracts.Record, key string) string {
	value, ok := record[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func normalizePagination(page int, perPage int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	return page, perPage
}
