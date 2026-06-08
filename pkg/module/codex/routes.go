package codex

type Route struct {
	Methods []string `json:"methods"`
	Path    string   `json:"path"`
}

type Discovery struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Routes         []Route  `json:"routes"`
	Authentication []string `json:"authentication"`
	Links          Links    `json:"links"`
}

type Links struct {
	Self      string `json:"self"`
	Namespace string `json:"namespace,omitempty"`
	Admin     string `json:"admin,omitempty"`
}

type ListEnvelope[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type ResourceEnvelope[T any] struct {
	Data T `json:"data"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Status    int            `json:"status"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

func RootDiscovery() Discovery {
	return Discovery{
		Name:           "GoCMS",
		Version:        "go-codex.v0.1",
		Authentication: []string{"browser_session", "app_token", "dev_bearer"},
		Links:          Links{Self: "/go-json", Namespace: "/go-json/go/v2/", Admin: "/go-admin"},
		Routes: []Route{
			{Methods: []string{"GET"}, Path: "/go-json"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/"},
		},
	}
}

func V2Discovery() Discovery {
	return Discovery{
		Name:           "GoCMS REST API",
		Version:        "go/v2",
		Authentication: []string{"browser_session", "app_token", "dev_bearer"},
		Links:          Links{Self: "/go-json/go/v2/", Admin: "/go-admin"},
		Routes: []Route{
			{Methods: []string{"GET", "POST"}, Path: "/go-json/go/v2/posts"},
			{Methods: []string{"GET", "PATCH", "DELETE"}, Path: "/go-json/go/v2/posts/{id}"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/posts/by-slug/{slug}"},
			{Methods: []string{"GET", "POST"}, Path: "/go-json/go/v2/pages"},
			{Methods: []string{"GET", "PATCH", "DELETE"}, Path: "/go-json/go/v2/pages/{id}"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/pages/by-slug/{slug}"},
			{Methods: []string{"GET", "POST"}, Path: "/go-json/go/v2/media"},
			{Methods: []string{"GET", "POST"}, Path: "/go-json/go/v2/taxonomies"},
			{Methods: []string{"GET", "POST"}, Path: "/go-json/go/v2/taxonomies/{type}/terms"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/authors/{id}"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/search"},
			{Methods: []string{"GET"}, Path: "/go-json/go/v2/settings"},
		},
	}
}

func EmptyList[T any]() ListEnvelope[T] {
	return ListEnvelope[T]{Data: []T{}, Pagination: Pagination{Page: 1, PerPage: 20, Total: 0, TotalPages: 0}}
}

func NotFound(message string) ErrorEnvelope {
	return ErrorEnvelope{Error: ErrorBody{Code: "not_found", Message: message, Status: 404}}
}
