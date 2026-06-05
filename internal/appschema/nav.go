package appschema

import (
	"slices"

	modulecms "github.com/fastygo/app-gocms/pkg/module"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/panel"
)

// NavItem describes one admin sidebar link assembled from module resources and special screens.
type NavItem struct {
	Href       string
	Label      string
	Order      int
	Capability contracts.CapabilityID
}

// NavItems returns capability-tagged admin navigation entries sorted by Order.
func (r *Registry) NavItems() []NavItem {
	if r == nil {
		return nil
	}
	items := make([]NavItem, 0, len(r.Assembly.Context.Resources)+3)
	for _, resource := range r.Assembly.Context.Resources {
		items = append(items, NavItem{
			Href:       resource.BasePath,
			Label:      resource.Label,
			Order:      resource.Navigation.Order,
			Capability: capabilityFor(resource, panel.OperationList),
		})
	}
	items = append(items, specialNavItems()...)
	slices.SortStableFunc(items, func(a, b NavItem) int {
		if a.Order == b.Order {
			return 0
		}
		if a.Order < b.Order {
			return -1
		}
		return 1
	})
	return items
}

func specialNavItems() []NavItem {
	return []NavItem{
		{Href: "/go-admin/menus", Label: "Menus", Order: 90, Capability: modulecms.CapabilitySettingsManage},
		{Href: "/go-admin/settings", Label: "Settings", Order: 91, Capability: modulecms.CapabilitySettingsManage},
		{Href: "/go-admin/import-export", Label: "Import / Export", Order: 92, Capability: modulecms.CapabilitySettingsManage},
	}
}
