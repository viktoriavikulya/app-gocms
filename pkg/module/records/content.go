package records

import "github.com/fastygo/platform/pkg/toolset"

const (
	RecordPost               toolset.RecordTypeID = "post"
	RecordPage               toolset.RecordTypeID = "page"
	RecordContentType        toolset.RecordTypeID = "content_type"
	RecordContentMeta        toolset.RecordTypeID = "content_meta_definition"
	RecordTaxonomy           toolset.RecordTypeID = "taxonomy"
	RecordTerm               toolset.RecordTypeID = "term"
	RecordMediaAsset         toolset.RecordTypeID = "media_asset"
	RecordAuthor             toolset.RecordTypeID = "author"
	CapabilityContentRead    toolset.CapabilityID = "content.read"
	CapabilityContentWrite   toolset.CapabilityID = "content.write" // deprecated compat alias
	CapabilityContentCreate  toolset.CapabilityID = "content.create"
	CapabilityContentEditOwn toolset.CapabilityID = "content.edit_own"
	CapabilityContentDelete  toolset.CapabilityID = "content.delete"
	CapabilityContentPrivate toolset.CapabilityID = "content.read_private"
	CapabilityMediaUpload    toolset.CapabilityID = "media.upload"
	CapabilityMediaEdit      toolset.CapabilityID = "media.edit"
	CapabilityTaxonomyManage toolset.CapabilityID = "taxonomies.manage"
	CapabilityTaxonomyAssign toolset.CapabilityID = "taxonomies.assign"
	CapabilityUsersManage    toolset.CapabilityID = "users.manage"
	CapabilitySettingsManage toolset.CapabilityID = "settings.manage"
)

func All() []toolset.RecordTypeDefinition {
	return []toolset.RecordTypeDefinition{
		Post(),
		Page(),
		ContentType(),
		ContentMetaDefinition(),
		Taxonomy(),
		Term(),
		MediaAsset(),
		Author(),
	}
}

func Post() toolset.RecordTypeDefinition {
	return contentRecord(RecordPost, "Post", "/posts/{slug}")
}

func Page() toolset.RecordTypeDefinition {
	return contentRecord(RecordPage, "Page", "/{slug}")
}

func ContentType() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordContentType,
		Label:         "Content Type",
		Description:   "Custom content type declaration compatible with GoCMS built-ins.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "id", Label: "ID", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "label", Label: "Label", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "supports", Label: "Supports", Type: toolset.FieldJSON, Description: "Editor features supported by this content type."},
			{ID: "permalink_pattern", Label: "Permalink pattern", Type: toolset.FieldText, Required: true},
			{ID: "public", Label: "Public", Type: toolset.FieldBoolean},
		},
		Capabilities: []toolset.CapabilityID{CapabilitySettingsManage},
	}
}

func ContentMetaDefinition() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordContentMeta,
		Label:         "Content Metadata Definition",
		Description:   "Custom field definition for CMS content metadata.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "key", Label: "Key", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "label", Label: "Label", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "scope", Label: "Scope", Type: toolset.FieldSelect, Required: true, Options: scopeOptions()},
			{ID: "kinds", Label: "Content kinds", Type: toolset.FieldJSON, Description: "Allowed content kinds, such as post or page."},
			{ID: "type", Label: "Field type", Type: toolset.FieldSelect, Required: true, Options: metaTypeOptions()},
			{ID: "public", Label: "Public", Type: toolset.FieldBoolean, Description: "Public metadata can be projected through /go-json."},
			{ID: "capability", Label: "Capability", Type: toolset.FieldText},
			{ID: "validation", Label: "Validation", Type: toolset.FieldJSON},
		},
		Capabilities: []toolset.CapabilityID{CapabilitySettingsManage},
	}
}

func Taxonomy() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordTaxonomy,
		Label:         "Taxonomy",
		Description:   "Taxonomy definition such as category or tag.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "type", Label: "Type", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "label", Label: "Label", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "mode", Label: "Mode", Type: toolset.FieldSelect, Required: true, Options: taxonomyModeOptions()},
			{ID: "assigned_kinds", Label: "Assigned content kinds", Type: toolset.FieldJSON},
			{ID: "public", Label: "Public", Type: toolset.FieldBoolean},
		},
		Capabilities: []toolset.CapabilityID{CapabilityTaxonomyManage},
	}
}

func Term() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordTerm,
		Label:         "Term",
		Description:   "Taxonomy term assigned to content.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "id", Label: "ID", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "taxonomy_type", Label: "Taxonomy", Type: toolset.FieldRelation, Required: true, Indexed: true},
			{ID: "name", Label: "Name", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "slug", Label: "Slug", Type: toolset.FieldText, Required: true, Searchable: true, Indexed: true},
			{ID: "parent_id", Label: "Parent term", Type: toolset.FieldRelation, Indexed: true},
			{ID: "description", Label: "Description", Type: toolset.FieldTextarea},
		},
		Capabilities: []toolset.CapabilityID{CapabilityTaxonomyManage, CapabilityTaxonomyAssign},
	}
}

func MediaAsset() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordMediaAsset,
		Label:         "Media Asset",
		Description:   "Metadata-only media asset. Binary storage is provided by adapters/plugins.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "id", Label: "ID", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "title", Label: "Title", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "mime_type", Label: "MIME type", Type: toolset.FieldText, Indexed: true},
			{ID: "public_url", Label: "Public URL", Type: toolset.FieldText},
			{ID: "provider_ref", Label: "Provider reference", Type: toolset.FieldText, Sensitive: true},
			{ID: "alt_text", Label: "Alt text", Type: toolset.FieldText, Searchable: true},
			{ID: "variants", Label: "Variants", Type: toolset.FieldJSON},
			{ID: "metadata", Label: "Metadata", Type: toolset.FieldJSON},
		},
		Capabilities: []toolset.CapabilityID{CapabilityMediaUpload, CapabilityMediaEdit},
	}
}

func Author() toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            RecordAuthor,
		Label:         "Author",
		Description:   "Public author profile projection, not private user credentials.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "id", Label: "ID", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "display_name", Label: "Display name", Type: toolset.FieldText, Required: true, Searchable: true},
			{ID: "slug", Label: "Slug", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "bio", Label: "Bio", Type: toolset.FieldTextarea},
			{ID: "avatar_media_id", Label: "Avatar media", Type: toolset.FieldRelation},
			{ID: "active", Label: "Active", Type: toolset.FieldBoolean, Indexed: true},
		},
		Capabilities: []toolset.CapabilityID{CapabilityContentPrivate, CapabilityUsersManage},
	}
}

func contentRecord(id toolset.RecordTypeID, label string, permalink string) toolset.RecordTypeDefinition {
	return toolset.RecordTypeDefinition{
		ID:            id,
		Label:         label,
		Description:   label + " content entry compatible with GoCMS content contracts.",
		SchemaVersion: "cms.k1",
		OwnerModule:   "cms",
		Scope:         toolset.ScopeWorkspace,
		Fields: []toolset.FieldDefinition{
			{ID: "id", Label: "ID", Type: toolset.FieldText, Required: true, Unique: true, Indexed: true},
			{ID: "title", Label: "Title", Type: toolset.FieldJSON, Required: true, Searchable: true, Description: "Localized title map."},
			{ID: "slug", Label: "Slug", Type: toolset.FieldJSON, Required: true, Searchable: true, Indexed: true, Description: "Localized slug map."},
			{ID: "content", Label: "Content", Type: toolset.FieldRichText, Searchable: true},
			{ID: "excerpt", Label: "Excerpt", Type: toolset.FieldTextarea, Searchable: true},
			{ID: "status", Label: "Status", Type: toolset.FieldSelect, Required: true, Indexed: true, Options: ContentStatusOptions()},
			{ID: "visibility", Label: "Visibility", Type: toolset.FieldSelect, Required: true, Indexed: true, Options: visibilityOptions()},
			{ID: "author_id", Label: "Author", Type: toolset.FieldRelation, Indexed: true},
			{ID: "featured_media_id", Label: "Featured media", Type: toolset.FieldRelation, Indexed: true},
			{ID: "taxonomy_term_ids", Label: "Categories", Type: toolset.FieldText, Description: "Comma-separated taxonomy term IDs."},
			{ID: "metadata", Label: "Metadata", Type: toolset.FieldJSON, Description: "Content metadata with public/private projection rules."},
			{ID: "permalink", Label: "Permalink", Type: toolset.FieldText, DefaultValue: permalink},
			{ID: "published_at", Label: "Published at", Type: toolset.FieldDateTime, Indexed: true},
			{ID: "scheduled_for", Label: "Scheduled for", Type: toolset.FieldDateTime, Indexed: true},
			{ID: "created_at", Label: "Created at", Type: toolset.FieldDateTime, Indexed: true},
			{ID: "updated_at", Label: "Updated at", Type: toolset.FieldDateTime, Indexed: true},
		},
		Capabilities: []toolset.CapabilityID{CapabilityContentRead, CapabilityContentCreate, CapabilityContentEditOwn, CapabilityContentDelete, CapabilityContentPrivate},
		Visibility:   "public",
	}
}

func ContentStatusOptions() []toolset.Option {
	return []toolset.Option{
		{Value: "draft", Label: "Draft"},
		{Value: "scheduled", Label: "Scheduled"},
		{Value: "published", Label: "Published"},
		{Value: "archived", Label: "Archived"},
		{Value: "trashed", Label: "Trashed"},
	}
}

func DefaultMetaDefinitions() []string {
	return []string{"seo_title", "seo_description", "seo_canonical_url", "seo_noindex"}
}

func visibilityOptions() []toolset.Option {
	return []toolset.Option{{Value: "public", Label: "Public"}, {Value: "private", Label: "Private"}, {Value: "password", Label: "Password protected"}}
}

func scopeOptions() []toolset.Option {
	return []toolset.Option{{Value: "content", Label: "Content"}}
}

func metaTypeOptions() []toolset.Option {
	return []toolset.Option{
		{Value: "text", Label: "Text"},
		{Value: "textarea", Label: "Textarea"},
		{Value: "boolean", Label: "Boolean"},
		{Value: "url", Label: "URL"},
		{Value: "json", Label: "JSON"},
	}
}

func taxonomyModeOptions() []toolset.Option {
	return []toolset.Option{{Value: "hierarchical", Label: "Hierarchical"}, {Value: "flat", Label: "Flat"}}
}
