package admin

// Action token scopes match pkg/app actionScopeForScreen minting (Phase 0).
const (
	ActionContentWrite     = "admin.content.write"
	ActionContentPublish   = "admin.content.publish"
	ActionContentSchedule  = "admin.content.schedule"
	ActionContentRestore   = "admin.content.restore"
	ActionContentArchive   = "admin.content.archive"
	ActionContentRevisions = "admin.content.revisions"
	ActionMediaUpload      = "admin.media.upload"
	ActionSettingsWrite    = "admin.settings.write"
	ActionMenusWrite       = "admin.menus.write"
)
