package modulecms

import "github.com/fastygo/platform/pkg/contracts"

const (
	CapabilityAdminAccess    contracts.CapabilityID = "admin.access"
	CapabilityContentRead    contracts.CapabilityID = "content.read"
	CapabilityContentWrite   contracts.CapabilityID = "content.write"
	CapabilityContentPrivate contracts.CapabilityID = "content.read_private"
	CapabilityMediaUpload    contracts.CapabilityID = "media.upload"
	CapabilityMediaEdit      contracts.CapabilityID = "media.edit"
	CapabilityTaxonomyManage contracts.CapabilityID = "taxonomies.manage"
	CapabilityTaxonomyAssign contracts.CapabilityID = "taxonomies.assign"
	CapabilityUsersManage    contracts.CapabilityID = "users.manage"
	CapabilitySettingsManage contracts.CapabilityID = "settings.manage"
)

func CapabilityDefinitions() []contracts.CapabilityDefinition {
	return []contracts.CapabilityDefinition{
		{ID: CapabilityAdminAccess, Label: "Access admin"},
		{ID: CapabilityContentRead, Label: "Read content"},
		{ID: CapabilityContentWrite, Label: "Write content"},
		{ID: CapabilityContentPrivate, Label: "Read private content"},
		{ID: CapabilityMediaUpload, Label: "Upload media"},
		{ID: CapabilityMediaEdit, Label: "Edit media"},
		{ID: CapabilityTaxonomyManage, Label: "Manage taxonomies"},
		{ID: CapabilityTaxonomyAssign, Label: "Assign taxonomies"},
		{ID: CapabilityUsersManage, Label: "Manage users"},
		{ID: CapabilitySettingsManage, Label: "Manage settings"},
	}
}
