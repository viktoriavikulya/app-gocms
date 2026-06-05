package operations

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

	"github.com/fastygo/platform/pkg/contracts"
)

type AuditRecorder interface {
	RecordAudit(event AuditEvent)
	Audit() []AuditEvent
}

type SQLiteAuditStore struct {
	db        *sql.DB
	workspace contracts.WorkspaceID
	mu        sync.Mutex
	limit     int
}

func NewSQLiteAuditStore(db *sql.DB, workspace contracts.WorkspaceID, limit int) *SQLiteAuditStore {
	if limit <= 0 {
		limit = 1000
	}
	if workspace == "" {
		workspace = "root"
	}
	return &SQLiteAuditStore{db: db, workspace: workspace, limit: limit}
}

func (s *SQLiteAuditStore) RecordAudit(event AuditEvent) {
	if s == nil || s.db == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	event.CreatedAt = time.Now().UTC()
	event.Details = redact(event.Details)
	detailsJSON, _ := json.Marshal(event.Details)
	_, _ = s.db.ExecContext(context.Background(), `INSERT INTO audit_events (workspace_id, id, actor_id, action, resource_type, resource_id, details_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		string(s.workspace),
		randomAuditID(),
		event.Actor,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		string(detailsJSON),
		event.CreatedAt.Format(time.RFC3339Nano),
	)
}

func (s *SQLiteAuditStore) Audit() []AuditEvent {
	if s == nil || s.db == nil {
		return nil
	}
	rows, err := s.db.QueryContext(context.Background(), `SELECT actor_id, action, resource_type, resource_id, details_json, created_at FROM audit_events WHERE workspace_id = ? ORDER BY created_at DESC LIMIT ?`, string(s.workspace), s.limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	events := []AuditEvent{}
	for rows.Next() {
		var actor, action, resourceType, resourceID, detailsJSON, createdAt string
		if err := rows.Scan(&actor, &action, &resourceType, &resourceID, &detailsJSON, &createdAt); err != nil {
			continue
		}
		event := AuditEvent{Actor: actor, Action: action, ResourceType: resourceType, ResourceID: resourceID}
		if resourceType != "" && resourceID != "" {
			event.Resource = resourceType + "/" + resourceID
		}
		_ = json.Unmarshal([]byte(detailsJSON), &event.Details)
		if parsed, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
			event.CreatedAt = parsed
		}
		events = append(events, event)
	}
	return events
}

func randomAuditID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
