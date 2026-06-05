package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fastygo/app-gocms/internal/application/wiring"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/pkg/module/capcheck"
	"github.com/fastygo/app-gocms/pkg/module/codex"
)

type restTransitionFn func(context.Context, wiring.Services, domaincontent.ID) (domaincontent.Entry, error)
type capCheckFn func(capcheck.Principal) bool

func (h Handler) restTransition(ctx context.Context, r *http.Request, s wiring.Services, kind domaincontent.Kind, id string, allowed capCheckFn, auditAction string, run restTransitionFn) (int, any) {
	principal, status, payload, ok := h.requirePrincipal(r)
	if !ok {
		return status, payload
	}
	if !allowed(principal) {
		return forbidden()
	}
	entry, found, err := s.Content.Get(ctx, domaincontent.ID(id))
	if err != nil {
		return serverError(err)
	}
	if !found || entry.Kind != kind {
		return notFound("content not found")
	}
	updated, err := run(ctx, s, domaincontent.ID(id))
	if err != nil {
		return errorResponse(err)
	}
	if h.Audit != nil {
		h.Audit.RecordAudit(operations.AuditEvent{Action: auditAction, Actor: string(principal.ID()), Resource: string(kind) + "/" + id, ResourceType: string(kind), ResourceID: id})
	}
	return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(updated, true)}
}

func (h Handler) restSchedule(ctx context.Context, r *http.Request, s wiring.Services, kind domaincontent.Kind, id string) (int, any) {
	principal, status, payload, ok := h.requirePrincipal(r)
	if !ok {
		return status, payload
	}
	if !capcheck.CanSchedule(principal) {
		return forbidden()
	}
	var body struct {
		PublishAt time.Time `json:"publish_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return validationError(err.Error())
	}
	updated, err := s.Content.Schedule(ctx, domaincontent.ID(id), body.PublishAt)
	if err != nil {
		return errorResponse(err)
	}
	if h.Audit != nil {
		h.Audit.RecordAudit(operations.AuditEvent{Action: operations.ActionRESTContentSchedule, Actor: string(principal.ID()), Resource: string(kind) + "/" + id, ResourceType: string(kind), ResourceID: id})
	}
	return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(updated, true)}
}

func (h Handler) contentRevisions(ctx context.Context, r *http.Request, readCtx requestContext, parts []string, s wiring.Services, kind domaincontent.Kind) (int, any) {
	if len(parts) < 2 {
		return notFound("route not found")
	}
	id := domaincontent.ID(parts[1])
	switch {
	case len(parts) == 3 && parts[2] == "revisions" && r.Method == http.MethodGet:
		principal, status, payload, ok := h.requirePrincipal(r)
		if !ok {
			return status, payload
		}
		if !capcheck.CanManageRevisions(principal) {
			return forbidden()
		}
		items, err := s.Revisions.ListByEntry(ctx, id, 50)
		if err != nil {
			return serverError(err)
		}
		return http.StatusOK, codex.ResourceEnvelope[[]RevisionDTO]{Data: projectRevisions(items, readCtx.includePrivate)}
	case len(parts) == 4 && parts[2] == "revisions" && r.Method == http.MethodGet:
		principal, status, payload, ok := h.requirePrincipal(r)
		if !ok {
			return status, payload
		}
		if !capcheck.CanManageRevisions(principal) {
			return forbidden()
		}
		revision, ok, err := s.Revisions.Get(ctx, parts[3])
		if err != nil {
			return serverError(err)
		}
		if !ok || revision.EntryID != id {
			return notFound("revision not found")
		}
		return http.StatusOK, codex.ResourceEnvelope[RevisionDTO]{Data: RevisionProjection(revision, readCtx.includePrivate)}
	case len(parts) == 5 && parts[2] == "revisions" && parts[4] == "restore" && r.Method == http.MethodPost:
		principal, status, payload, ok := h.requirePrincipal(r)
		if !ok {
			return status, payload
		}
		if !capcheck.CanManageRevisions(principal) {
			return forbidden()
		}
		restored, err := s.Revisions.Restore(ctx, parts[3])
		if err != nil {
			return errorResponse(err)
		}
		if err := s.Content.Update(ctx, restored); err != nil {
			return errorResponse(err)
		}
		if h.Audit != nil {
			h.Audit.RecordAudit(operations.AuditEvent{Action: operations.ActionRESTContentRevision, Actor: string(principal.ID()), Resource: string(kind) + "/" + string(id), ResourceType: string(kind), ResourceID: string(id), Details: map[string]any{"revision_id": parts[3]}})
		}
		return http.StatusOK, codex.ResourceEnvelope[ContentDTO]{Data: ContentProjection(restored, true)}
	default:
		return notFound("route not found")
	}
}
