package modulecms

import (
	modulecaps "github.com/fastygo/app-gocms/pkg/module/capabilities"
	"github.com/fastygo/platform/pkg/contracts"
)

const (
	CapabilityAdminAccess    = modulecaps.AdminAccess
	CapabilityContentRead    = modulecaps.ContentRead
	CapabilityContentWrite   = modulecaps.ContentWrite
	CapabilityContentPrivate = modulecaps.ContentPrivate
	CapabilityMediaUpload    = modulecaps.MediaUpload
	CapabilityMediaEdit      = modulecaps.MediaEdit
	CapabilityTaxonomyManage = modulecaps.TaxonomyManage
	CapabilityTaxonomyAssign = modulecaps.TaxonomyAssign
	CapabilityUsersManage    = modulecaps.UsersManage
	CapabilitySettingsManage = modulecaps.SettingsManage
)

func CapabilityDefinitions() []contracts.CapabilityDefinition {
	return modulecaps.Definitions()
}
