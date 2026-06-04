package settings

type Value struct {
	Key     string `json:"key"`
	Group   string `json:"group"`
	Value   any    `json:"value"`
	Public  bool   `json:"public"`
	Private bool   `json:"private"`
}

type Definition struct {
	Key          string `json:"key"`
	Group        string `json:"group"`
	DefaultValue any    `json:"default_value"`
	Public       bool   `json:"public"`
}
