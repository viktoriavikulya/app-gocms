package media

type Asset struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	MIMEType    string         `json:"mime_type"`
	PublicURL   string         `json:"public_url"`
	ProviderRef string         `json:"provider_ref"`
	AltText     string         `json:"alt_text"`
	Variants    map[string]any `json:"variants"`
	Metadata    map[string]any `json:"metadata"`
}

type BlobStore interface {
	PublicURL(providerRef string) string
}
