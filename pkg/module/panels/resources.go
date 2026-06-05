package panels

import (
	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/contracts"
	"github.com/fastygo/platform/pkg/panel"
	"github.com/fastygo/platform/pkg/panelschema"
	"github.com/fastygo/platform/pkg/toolset"
)

const (
	CapabilityContentRead            contracts.CapabilityID = "content.read"
	CapabilityContentCreate          contracts.CapabilityID = "content.create"
	CapabilityContentEditOwn         contracts.CapabilityID = "content.edit_own"
	CapabilityContentPublish         contracts.CapabilityID = "content.publish"
	CapabilityContentSchedule        contracts.CapabilityID = "content.schedule"
	CapabilityContentDelete          contracts.CapabilityID = "content.delete"
	CapabilityContentPrivate         contracts.CapabilityID = "content.read_private"
	CapabilityMediaUpload            contracts.CapabilityID = "media.upload"
	CapabilityTaxonomyManage         contracts.CapabilityID = "taxonomies.manage"
	CapabilityTaxonomyAssign         contracts.CapabilityID = "taxonomies.assign"
	CapabilityUsersManage            contracts.CapabilityID = "users.manage"
	CapabilitySettingsManage         contracts.CapabilityID = "settings.manage"
)

func Resources() []ResourceBinding {
	return []ResourceBinding{
		contentResource(records.Post(), "/go-admin/posts", "file", 1),
		contentResource(records.Page(), "/go-admin/pages", "book", 2),
		resource(records.ContentType(), "/go-admin/content-types", "blocks", 3, CapabilitySettingsManage, CapabilitySettingsManage),
		resource(records.Taxonomy(), "/go-admin/taxonomies", "boxes", 4, CapabilityTaxonomyManage, CapabilityTaxonomyManage),
		resource(records.Term(), "/go-admin/terms", "tag", 5, CapabilityTaxonomyManage, CapabilityTaxonomyManage),
		resource(records.ContentMetaDefinition(), "/go-admin/meta", "settings", 6, CapabilitySettingsManage, CapabilitySettingsManage),
		resource(records.MediaAsset(), "/go-admin/media", "image", 7, CapabilityContentRead, CapabilityMediaUpload),
		resource(records.Author(), "/go-admin/authors", "user", 8, CapabilityContentPrivate, CapabilityUsersManage),
	}
}

func Views() []panelschema.ViewDescriptor {
	result := []panelschema.ViewDescriptor{}
	for _, resource := range Resources() {
		result = append(result, panelschema.TableView(string(resource.Resource.ID)+"-table", resource.Resource.Label, resource.Resource, resource.Record))
	}
	return result
}

func Workflows() []panelschema.WorkflowDescriptor {
	return []panelschema.WorkflowDescriptor{
		panelschema.ContentStatusWorkflowFor(records.RecordPost, CapabilityContentPublish),
		panelschema.ContentStatusWorkflowFor(records.RecordPage, CapabilityContentPublish),
	}
}

func Actions() []panelschema.ActionDescriptor {
	return []panelschema.ActionDescriptor{
		panelschema.WorkflowAction("publish-post", "Publish", "post-status", "publish", CapabilityContentPublish),
		panelschema.WorkflowAction("schedule-post", "Schedule", "post-status", "schedule", CapabilityContentSchedule),
		panelschema.WorkflowAction("trash-post", "Trash", "post-status", "trash", CapabilityContentDelete),
	}
}

func RelationViews() []panelschema.RelationViewDescriptor {
	return []panelschema.RelationViewDescriptor{
		panelschema.RelationViewFromDefinition(relation("post_terms", records.RecordPost, records.RecordTerm, toolset.RelationManyToMany), panelschema.RelationPicker, CapabilityTaxonomyAssign),
		panelschema.RelationViewFromDefinition(relation("page_terms", records.RecordPage, records.RecordTerm, toolset.RelationManyToMany), panelschema.RelationPicker, CapabilityTaxonomyAssign),
		panelschema.RelationViewFromDefinition(relation("post_author", records.RecordPost, records.RecordAuthor, toolset.RelationOneToOne), panelschema.RelationPicker, CapabilityContentEditOwn),
		panelschema.RelationViewFromDefinition(relation("post_featured_media", records.RecordPost, records.RecordMediaAsset, toolset.RelationOneToOne), panelschema.RelationPicker, CapabilityContentEditOwn),
	}
}

type ResourceBinding struct {
	Record   toolset.RecordTypeDefinition
	Resource panel.Resource[contracts.CapabilityID]
}

func contentResource(record toolset.RecordTypeDefinition, basePath string, icon string, order int) ResourceBinding {
	return ResourceBinding{
		Record: record,
		Resource: panelschema.MustResourceFromRecord(record, panelschema.ResourceOptions{
			BasePath: basePath,
			Icon:     icon,
			Order:    order,
			List:     CapabilityContentRead,
			Create:   CapabilityContentCreate,
			Edit:     CapabilityContentEditOwn,
			Delete:   CapabilityContentDelete,
		}),
	}
}

func resource(record toolset.RecordTypeDefinition, basePath string, icon string, order int, read contracts.CapabilityID, write contracts.CapabilityID) ResourceBinding {
	return ResourceBinding{
		Record: record,
		Resource: panelschema.MustResourceFromRecord(record, panelschema.ResourceOptions{
			BasePath: basePath,
			Icon:     icon,
			Order:    order,
			List:     read,
			Create:   write,
			Edit:     write,
			Delete:   write,
		}),
	}
}

func relation(id toolset.RelationID, source toolset.RecordTypeID, target toolset.RecordTypeID, cardinality toolset.RelationCardinality) toolset.RelationDefinition {
	return toolset.RelationDefinition{
		ID:          id,
		Source:      source,
		Target:      target,
		Cardinality: cardinality,
		Policy:      toolset.RelationPolicy{CrossWorkspaceMode: toolset.CrossWorkspaceForbidden},
	}
}
