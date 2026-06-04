package migration

type Mapping struct {
	SourcePath string
	Target     string
	Notes      string
}

func CurrentGoCMSMappings() []Mapping {
	return []Mapping{
		{SourcePath: "internal/domain/content/content.go", Target: "records.Post/Page", Notes: "Entry lifecycle, localized fields, status, visibility, metadata."},
		{SourcePath: "internal/domain/contenttype/contenttype.go", Target: "records.ContentType", Notes: "Built-in post/page supports and permalink patterns."},
		{SourcePath: "internal/domain/meta/meta.go", Target: "records.ContentMetaDefinition", Notes: "Custom metadata definitions and public projection rules."},
		{SourcePath: "internal/domain/taxonomy/taxonomy.go", Target: "records.Taxonomy/Term + relations", Notes: "Category/tag definitions and content assignment."},
		{SourcePath: "internal/domain/media/media.go", Target: "records.MediaAsset", Notes: "Metadata-only media core; binary storage stays adapter-owned."},
		{SourcePath: "internal/domain/users/users.go", Target: "records.Author", Notes: "Public author profile projection only."},
	}
}
