package modulecms

import "github.com/fastygo/platform/pkg/contracts"

const (
	CapabilityAdminAccess    contracts.CapabilityID = "admin.access"
	CapabilityContentRead    contracts.CapabilityID = "content.read"
	CapabilityContentWrite   contracts.CapabilityID = "content.write" // deprecated compat alias
	CapabilityContentPrivate contracts.CapabilityID = "content.read_private"

	CapabilityContentCreate          contracts.CapabilityID = "content.create"
	CapabilityContentEdit            contracts.CapabilityID = "content.edit"
	CapabilityContentEditOwn         contracts.CapabilityID = "content.edit_own"
	CapabilityContentEditOthers      contracts.CapabilityID = "content.edit_others"
	CapabilityContentPublish         contracts.CapabilityID = "content.publish"
	CapabilityContentSchedule        contracts.CapabilityID = "content.schedule"
	CapabilityContentArchive         contracts.CapabilityID = "content.archive"
	CapabilityContentDelete          contracts.CapabilityID = "content.delete"
	CapabilityContentRestore         contracts.CapabilityID = "content.restore"
	CapabilityContentManageRevisions contracts.CapabilityID = "content.manage_revisions"

	CapabilityMediaUpload     contracts.CapabilityID = "media.upload"
	CapabilityMediaEdit       contracts.CapabilityID = "media.edit"
	CapabilityMediaDelete     contracts.CapabilityID = "media.delete"
	CapabilityMediaReadPrivate contracts.CapabilityID = "media.read_private"

	CapabilityTaxonomyManage contracts.CapabilityID = "taxonomies.manage"
	CapabilityTaxonomyAssign contracts.CapabilityID = "taxonomies.assign"

	CapabilityThemesView           contracts.CapabilityID = "themes.view"
	CapabilityThemesActivate       contracts.CapabilityID = "themes.activate"
	CapabilityThemesManageSettings contracts.CapabilityID = "themes.manage_settings"

	CapabilityPluginsView           contracts.CapabilityID = "plugins.view"
	CapabilityPluginsInstall        contracts.CapabilityID = "plugins.install"
	CapabilityPluginsActivate       contracts.CapabilityID = "plugins.activate"
	CapabilityPluginsDeactivate     contracts.CapabilityID = "plugins.deactivate"
	CapabilityPluginsUninstall      contracts.CapabilityID = "plugins.uninstall"
	CapabilityPluginsManageSettings contracts.CapabilityID = "plugins.manage_settings"

	CapabilityUsersView   contracts.CapabilityID = "users.view"
	CapabilityUsersCreate contracts.CapabilityID = "users.create"
	CapabilityUsersEdit   contracts.CapabilityID = "users.edit"
	CapabilityUsersDelete contracts.CapabilityID = "users.delete"
	CapabilityUsersManage contracts.CapabilityID = "users.manage" // compat alias

	CapabilityRolesView   contracts.CapabilityID = "roles.view"
	CapabilityRolesManage contracts.CapabilityID = "roles.manage"

	CapabilitySettingsView   contracts.CapabilityID = "settings.view"
	CapabilitySettingsManage contracts.CapabilityID = "settings.manage"

	CapabilityRESTAccess        contracts.CapabilityID = "rest.access"
	CapabilityRESTAccessPrivate contracts.CapabilityID = "rest.access_private"
	CapabilityRESTWrite         contracts.CapabilityID = "rest.write"
)

func CapabilityDefinitions() []contracts.CapabilityDefinition {
	return []contracts.CapabilityDefinition{
		{ID: CapabilityAdminAccess, Label: "Access admin"},
		{ID: CapabilityContentRead, Label: "Read content"},
		{ID: CapabilityContentWrite, Label: "Write content (compat)"},
		{ID: CapabilityContentPrivate, Label: "Read private content"},
		{ID: CapabilityContentCreate, Label: "Create content"},
		{ID: CapabilityContentEdit, Label: "Edit any content"},
		{ID: CapabilityContentEditOwn, Label: "Edit own content"},
		{ID: CapabilityContentEditOthers, Label: "Edit others content"},
		{ID: CapabilityContentPublish, Label: "Publish content"},
		{ID: CapabilityContentSchedule, Label: "Schedule content"},
		{ID: CapabilityContentArchive, Label: "Archive content"},
		{ID: CapabilityContentDelete, Label: "Delete content"},
		{ID: CapabilityContentRestore, Label: "Restore content"},
		{ID: CapabilityContentManageRevisions, Label: "Manage revisions"},
		{ID: CapabilityMediaUpload, Label: "Upload media"},
		{ID: CapabilityMediaEdit, Label: "Edit media"},
		{ID: CapabilityMediaDelete, Label: "Delete media"},
		{ID: CapabilityMediaReadPrivate, Label: "Read private media"},
		{ID: CapabilityTaxonomyManage, Label: "Manage taxonomies"},
		{ID: CapabilityTaxonomyAssign, Label: "Assign taxonomies"},
		{ID: CapabilityThemesView, Label: "View themes"},
		{ID: CapabilityThemesActivate, Label: "Activate themes"},
		{ID: CapabilityThemesManageSettings, Label: "Manage theme settings"},
		{ID: CapabilityPluginsView, Label: "View plugins"},
		{ID: CapabilityPluginsInstall, Label: "Install plugins"},
		{ID: CapabilityPluginsActivate, Label: "Activate plugins"},
		{ID: CapabilityPluginsDeactivate, Label: "Deactivate plugins"},
		{ID: CapabilityPluginsUninstall, Label: "Uninstall plugins"},
		{ID: CapabilityPluginsManageSettings, Label: "Manage plugin settings"},
		{ID: CapabilityUsersView, Label: "View users"},
		{ID: CapabilityUsersCreate, Label: "Create users"},
		{ID: CapabilityUsersEdit, Label: "Edit users"},
		{ID: CapabilityUsersDelete, Label: "Delete users"},
		{ID: CapabilityUsersManage, Label: "Manage users"},
		{ID: CapabilityRolesView, Label: "View roles"},
		{ID: CapabilityRolesManage, Label: "Manage roles"},
		{ID: CapabilitySettingsView, Label: "View settings"},
		{ID: CapabilitySettingsManage, Label: "Manage settings"},
		{ID: CapabilityRESTAccess, Label: "REST access"},
		{ID: CapabilityRESTAccessPrivate, Label: "REST private access"},
		{ID: CapabilityRESTWrite, Label: "REST write"},
	}
}
