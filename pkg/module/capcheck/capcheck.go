package capcheck

import (
	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/contracts"
)

type Principal interface {
	ID() contracts.PrincipalID
	Has(capability contracts.CapabilityID) bool
}

func HasWriteCompat(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentWrite) ||
		principal.Has(modulecms.CapabilityContentCreate) ||
		principal.Has(modulecms.CapabilityContentEdit) ||
		principal.Has(modulecms.CapabilityContentEditOwn) ||
		principal.Has(modulecms.CapabilityContentEditOthers) ||
		principal.Has(modulecms.CapabilityContentPublish) ||
		principal.Has(modulecms.CapabilityContentDelete)
}

func CanCreate(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentCreate) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanEdit(principal Principal, authorID string) bool {
	if principal.Has(modulecms.CapabilityContentEdit) || principal.Has(modulecms.CapabilityContentEditOthers) {
		return true
	}
	if principal.Has(modulecms.CapabilityContentEditOwn) && authorID == string(principal.ID()) {
		return true
	}
	return principal.Has(modulecms.CapabilityContentWrite)
}

func CanPublish(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentPublish) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanSchedule(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentSchedule) || principal.Has(modulecms.CapabilityContentPublish) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanDelete(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentDelete) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanRestore(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentRestore) || principal.Has(modulecms.CapabilityContentDelete) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanArchive(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentArchive) || principal.Has(modulecms.CapabilityContentDelete) || principal.Has(modulecms.CapabilityContentWrite)
}

func CanManageRevisions(principal Principal) bool {
	return principal.Has(modulecms.CapabilityContentManageRevisions) || principal.Has(modulecms.CapabilityContentEdit) || principal.Has(modulecms.CapabilityContentWrite)
}
