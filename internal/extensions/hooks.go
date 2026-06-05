package extensions

import (
	"time"

	"github.com/fastygo/platform/pkg/contracts"
)

const (
	HookContentCreateBefore  = "content.create.before"
	HookContentCreateAfter   = "content.create.after"
	HookContentUpdateBefore  = "content.update.before"
	HookContentUpdateAfter   = "content.update.after"
	HookContentStatusBefore  = "content.status.before"
	HookContentStatusAfter   = "content.status.after"
	HookContentTrashBefore   = "content.trash.before"
	HookContentTrashAfter    = "content.trash.after"
	HookContentRestoreBefore = "content.restore.before"
	HookContentRestoreAfter  = "content.restore.after"
	HookMediaUploadAfter     = "media.upload.after"
	HookMediaDeleteBefore    = "media.delete.before"
	HookMediaDeleteAfter     = "media.delete.after"
	HookSettingsUpdateBefore = "settings.update.before"
	HookSettingsUpdateAfter  = "settings.update.after"
)

type HookPayload struct {
	Hook        string
	PrincipalID contracts.PrincipalID
	RequestID   string
	Locale      string
	Entity      any
	Phase       string
	OccurredAt  time.Time
}
