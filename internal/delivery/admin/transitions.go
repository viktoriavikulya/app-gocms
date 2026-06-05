package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fastygo/app-gocms/internal/application/wiring"
	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/operations"
	"github.com/fastygo/app-gocms/pkg/module/capcheck"
)

func (h Handler) registerTransitions(mux *http.ServeMux, kind domaincontent.Kind, collection string) {
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/publish", h.contentTransition(kind, collection, capcheck.CanPublish, ActionContentPublish, operations.ActionAdminContentPublish, func(ctx context.Context, s wiring.Services, id domaincontent.ID) error {
		_, err := s.Content.Publish(ctx, id)
		return err
	}))
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/unpublish", h.contentTransition(kind, collection, capcheck.CanPublish, ActionContentPublish, operations.ActionAdminContentUnpublish, func(ctx context.Context, s wiring.Services, id domaincontent.ID) error {
		_, err := s.Content.Unpublish(ctx, id)
		return err
	}))
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/schedule", h.contentSchedule(kind, collection))
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/restore", h.contentTransition(kind, collection, capcheck.CanRestore, ActionContentRestore, operations.ActionAdminContentRestore, func(ctx context.Context, s wiring.Services, id domaincontent.ID) error {
		_, err := s.Content.Restore(ctx, id)
		return err
	}))
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/archive", h.contentTransition(kind, collection, capcheck.CanArchive, ActionContentArchive, operations.ActionAdminContentArchive, func(ctx context.Context, s wiring.Services, id domaincontent.ID) error {
		_, err := s.Content.Archive(ctx, id)
		return err
	}))
	mux.HandleFunc("POST /go-admin/"+collection+"/{id}/revisions/{revID}/restore", h.revisionRestore(kind, collection))
}

type transitionFn func(ctx context.Context, s wiring.Services, id domaincontent.ID) error
type capFn func(capcheck.Principal) bool

func (h Handler) contentTransition(kind domaincontent.Kind, collection string, allowed capFn, scope, auditAction string, run transitionFn) http.HandlerFunc {
	listPath := "/go-admin/" + collection
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !allowed(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, scope) {
			return
		}
		id := domaincontent.ID(r.PathValue("id"))
		err := h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			entry, found, err := s.Content.Get(ctx, id)
			if err != nil {
				return err
			}
			if !found || entry.Kind != kind {
				return errNotFound("content not found")
			}
			return run(ctx, s, id)
		})
		if err != nil {
			status := http.StatusBadRequest
			if isNotFound(err) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}
		h.recordAudit(principal, auditAction, string(kind), string(id), nil)
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func (h Handler) contentSchedule(kind domaincontent.Kind, collection string) http.HandlerFunc {
	listPath := "/go-admin/" + collection
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !capcheck.CanSchedule(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, ActionContentSchedule) {
			return
		}
		when := firstFormValue(r, "scheduled_at", "publish_at")
		at, err := time.Parse("2006-01-02T15:04", when)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid schedule time: %v", err), http.StatusBadRequest)
			return
		}
		id := domaincontent.ID(r.PathValue("id"))
		err = h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			entry, found, err := s.Content.Get(ctx, id)
			if err != nil {
				return err
			}
			if !found || entry.Kind != kind {
				return errNotFound("content not found")
			}
			_, err = s.Content.Schedule(ctx, id, at)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.recordAudit(principal, operations.ActionAdminContentSchedule, string(kind), string(id), map[string]any{"publish_at": at})
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func (h Handler) revisionRestore(kind domaincontent.Kind, collection string) http.HandlerFunc {
	listPath := "/go-admin/" + collection
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := h.requirePrincipal(w, r)
		if !ok {
			return
		}
		if !capcheck.CanManageRevisions(principal) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if !h.validateToken(w, r, ActionContentRevisions) {
			return
		}
		contentID := domaincontent.ID(r.PathValue("id"))
		revID := r.PathValue("revID")
		err := h.withWiring(r, func(ctx context.Context, s wiring.Services) error {
			entry, found, err := s.Content.Get(ctx, contentID)
			if err != nil {
				return err
			}
			if !found || entry.Kind != kind {
				return errNotFound("content not found")
			}
			restored, err := s.Revisions.Restore(ctx, revID)
			if err != nil {
				return err
			}
			return s.Content.Update(ctx, restored)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.recordAudit(principal, operations.ActionAdminContentRevision, string(kind), string(contentID), map[string]any{"revision_id": revID})
		http.Redirect(w, r, listPath, http.StatusSeeOther)
	}
}

func firstFormValue(r *http.Request, keys ...string) string {
	for _, key := range keys {
		if v := r.PostForm.Get(key); v != "" {
			return v
		}
	}
	return ""
}
