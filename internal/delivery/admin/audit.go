package admin

import (
	appauthn "github.com/fastygo/app-gocms/internal/application/authn"
	"github.com/fastygo/app-gocms/internal/operations"
)

func (h Handler) recordAudit(principal appauthn.Principal, action, resourceType, resourceID string, details map[string]any) {
	if h.Audit == nil {
		return
	}
	h.Audit.RecordAudit(operations.AuditEvent{
		Action:       action,
		Actor:          string(principal.ID()),
		Resource:       resourceType + "/" + resourceID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	})
}
