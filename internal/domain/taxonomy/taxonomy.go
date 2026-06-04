package taxonomy

type Mode string

const (
	ModeHierarchical Mode = "hierarchical"
	ModeFlat         Mode = "flat"
)

type Definition struct {
	Type          string   `json:"type"`
	Label         string   `json:"label"`
	Mode          Mode     `json:"mode"`
	AssignedKinds []string `json:"assigned_kinds"`
	Public        bool     `json:"public"`
}

type Term struct {
	ID           string `json:"id"`
	TaxonomyType string `json:"taxonomy_type"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ParentID     string `json:"parent_id"`
	Description  string `json:"description"`
}
