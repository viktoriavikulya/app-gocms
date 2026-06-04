package relations

import (
	"github.com/fastygo/app-gocms/pkg/module/records"
	"github.com/fastygo/platform/pkg/toolset"
)

func All() []toolset.RelationDefinition {
	return []toolset.RelationDefinition{
		TaxonomyTerms(),
		PostTerms(),
		PageTerms(),
		PostFeaturedMedia(),
		PageFeaturedMedia(),
		PostAuthor(),
		PageAuthor(),
		AuthorAvatar(),
	}
}

func TaxonomyTerms() toolset.RelationDefinition {
	return relation("taxonomy_terms", "Taxonomy terms", records.RecordTaxonomy, records.RecordTerm, toolset.RelationOneToMany, toolset.DeleteRestrict)
}

func PostTerms() toolset.RelationDefinition {
	return relation("post_terms", "Post terms", records.RecordPost, records.RecordTerm, toolset.RelationManyToMany, toolset.DeleteRestrict)
}

func PageTerms() toolset.RelationDefinition {
	return relation("page_terms", "Page terms", records.RecordPage, records.RecordTerm, toolset.RelationManyToMany, toolset.DeleteRestrict)
}

func PostFeaturedMedia() toolset.RelationDefinition {
	return relation("post_featured_media", "Post featured media", records.RecordPost, records.RecordMediaAsset, toolset.RelationOneToOne, toolset.DeleteNullify)
}

func PageFeaturedMedia() toolset.RelationDefinition {
	return relation("page_featured_media", "Page featured media", records.RecordPage, records.RecordMediaAsset, toolset.RelationOneToOne, toolset.DeleteNullify)
}

func PostAuthor() toolset.RelationDefinition {
	return relation("post_author", "Post author", records.RecordPost, records.RecordAuthor, toolset.RelationOneToOne, toolset.DeleteRestrict)
}

func PageAuthor() toolset.RelationDefinition {
	return relation("page_author", "Page author", records.RecordPage, records.RecordAuthor, toolset.RelationOneToOne, toolset.DeleteRestrict)
}

func AuthorAvatar() toolset.RelationDefinition {
	return relation("author_avatar", "Author avatar", records.RecordAuthor, records.RecordMediaAsset, toolset.RelationOneToOne, toolset.DeleteNullify)
}

func relation(id toolset.RelationID, label string, source toolset.RecordTypeID, target toolset.RecordTypeID, cardinality toolset.RelationCardinality, deleteBehavior toolset.DeleteBehavior) toolset.RelationDefinition {
	return toolset.RelationDefinition{
		ID:             id,
		Label:          label,
		Source:         source,
		Target:         target,
		Cardinality:    cardinality,
		DeleteBehavior: deleteBehavior,
		Policy: toolset.RelationPolicy{
			CrossWorkspaceMode: toolset.CrossWorkspaceForbidden,
		},
	}
}
