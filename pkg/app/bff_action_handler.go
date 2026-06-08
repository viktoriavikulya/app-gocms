package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/fastygo/platform/pkg/bff"
)

func (a authBoundary) handleBFFAction(executors *bff.ExecutorRegistry) http.HandlerFunc {
	return a.requireBFFJSON(func(w http.ResponseWriter, r *http.Request) {
		actionID := strings.TrimSpace(r.PathValue("actionId"))
		principal, _ := a.principalFromRequest(r)
		req, err := parseActionRequest(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		_, result, err := executors.SecureDispatch(r.Context(), bff.SecureDispatchInput{
			ActionID:    actionID,
			Request:     req,
			Principal:   principal,
			Actor:       principal.ID(),
			Validator:   a,
			Auditor:     cmsActionAudit(),
			ModuleID:    "cms",
			WorkspaceID: cmsWorkspace,
			Resource:    cmsActionResource(actionID),
		})
		switch {
		case errors.Is(err, bff.ErrUnknownAction):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "action not found"})
			return
		case errors.Is(err, bff.ErrForbidden):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
			return
		case errors.Is(err, bff.ErrInvalidActionToken):
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid action token"})
			return
		case err != nil:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, bff.ActionHTTPStatus(result, nil), result)
	})
}

func cmsActionResource(actionID string) string {
	if strings.HasPrefix(actionID, "post.") {
		return "post"
	}
	return "cms"
}
