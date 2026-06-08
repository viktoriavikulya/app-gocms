package capabilities

import "github.com/fastygo/platform/pkg/contracts"

const (
	AdminAccess    contracts.CapabilityID = "admin.access"
	ContentRead    contracts.CapabilityID = "content.read"
	ContentWrite   contracts.CapabilityID = "content.write"
	ContentPrivate contracts.CapabilityID = "content.read_private"
	MediaUpload    contracts.CapabilityID = "media.upload"
	MediaEdit      contracts.CapabilityID = "media.edit"
	TaxonomyManage contracts.CapabilityID = "taxonomies.manage"
	TaxonomyAssign contracts.CapabilityID = "taxonomies.assign"
	UsersManage    contracts.CapabilityID = "users.manage"
	SettingsManage contracts.CapabilityID = "settings.manage"
)

func Definitions() []contracts.CapabilityDefinition {
	return []contracts.CapabilityDefinition{
		{ID: AdminAccess, Label: "Access admin"},
		{ID: ContentRead, Label: "Read content"},
		{ID: ContentWrite, Label: "Write content"},
		{ID: ContentPrivate, Label: "Read private content"},
		{ID: MediaUpload, Label: "Upload media"},
		{ID: MediaEdit, Label: "Edit media"},
		{ID: TaxonomyManage, Label: "Manage taxonomies"},
		{ID: TaxonomyAssign, Label: "Assign taxonomies"},
		{ID: UsersManage, Label: "Manage users"},
		{ID: SettingsManage, Label: "Manage settings"},
	}
}
