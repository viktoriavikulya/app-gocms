package menus

type Item struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	ParentID string `json:"parent_id"`
}

type Menu struct {
	ID       string `json:"id"`
	Location string `json:"location"`
	Items    []Item `json:"items"`
}
