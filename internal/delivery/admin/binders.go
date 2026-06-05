package admin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
)

func bindContent(form url.Values, kind domaincontent.Kind, id domaincontent.ID) (domaincontent.Entry, error) {
	if id == "" {
		id = domaincontent.ID(strings.TrimSpace(form.Get("id")))
	}
	if id == "" {
		return domaincontent.Entry{}, fmt.Errorf("content id is required")
	}
	entry := domaincontent.Entry{
		ID:              id,
		Kind:            kind,
		Title:           map[string]string{"en": form.Get("title")},
		Slug:            form.Get("slug"),
		Content:         form.Get("content"),
		Excerpt:         form.Get("excerpt"),
		AuthorID:        defaultIfBlank(form.Get("author_id"), "admin"),
		FeaturedMediaID: form.Get("featured_media_id"),
		Status:          domaincontent.Status(defaultIfBlank(form.Get("status"), string(domaincontent.StatusDraft))),
		Visibility:      domaincontent.Visibility(defaultIfBlank(form.Get("visibility"), string(domaincontent.VisibilityPublic))),
	}
	if raw := strings.TrimSpace(form.Get("metadata")); raw != "" {
		entry.Metadata = map[string]any{}
		if err := json.Unmarshal([]byte(raw), &entry.Metadata); err != nil {
			return domaincontent.Entry{}, fmt.Errorf("invalid metadata json")
		}
	}
	if terms := form["taxonomy_term_ids"]; len(terms) > 0 {
		entry.TermIDs = append([]string{}, terms...)
	} else if raw := form.Get("taxonomy_term_ids"); raw != "" {
		entry.TermIDs = splitCSV(raw)
	} else if raw := form.Get("term_ids"); raw != "" {
		entry.TermIDs = splitCSV(raw)
	}
	if raw := strings.TrimSpace(firstValue(form, "scheduled_at", "scheduled_for", "publish_at")); raw != "" {
		if when, err := time.Parse("2006-01-02T15:04", raw); err == nil {
			entry.ScheduledFor = when.UTC()
		}
	}
	return entry, nil
}

func firstValue(form url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(form.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func bindTaxonomy(form url.Values) (taxonomy.Definition, error) {
	taxonomyType := strings.TrimSpace(form.Get("type"))
	if taxonomyType == "" {
		return taxonomy.Definition{}, fmt.Errorf("taxonomy type is required")
	}
	mode := taxonomy.Mode(defaultIfBlank(form.Get("mode"), string(taxonomy.ModeFlat)))
	return taxonomy.Definition{
		Type:  taxonomyType,
		Label: form.Get("label"),
		Mode:  mode,
		Public: form.Get("public") == "true" || form.Get("public") == "on" || form.Get("public") == "1",
	}, nil
}

func bindTerm(form url.Values, taxonomyType string, id string) (taxonomy.Term, error) {
	if id == "" {
		id = strings.TrimSpace(form.Get("id"))
	}
	if id == "" {
		return taxonomy.Term{}, fmt.Errorf("term id is required")
	}
	if taxonomyType == "" {
		taxonomyType = strings.TrimSpace(form.Get("taxonomy_type"))
	}
	if taxonomyType == "" {
		return taxonomy.Term{}, fmt.Errorf("taxonomy type is required")
	}
	return taxonomy.Term{
		ID:           id,
		TaxonomyType: taxonomyType,
		Name:         form.Get("name"),
		Slug:         form.Get("slug"),
		ParentID:     form.Get("parent_id"),
		Description:  form.Get("description"),
	}, nil
}

func bindMediaAsset(form url.Values, id string) (media.Asset, error) {
	if id == "" {
		id = strings.TrimSpace(form.Get("id"))
	}
	if id == "" {
		return media.Asset{}, fmt.Errorf("media id is required")
	}
	filename := firstNonEmpty(form.Get("filename"), form.Get("title"))
	asset := media.Asset{
		ID:          id,
		Title:       filename,
		MIMEType:    form.Get("mime_type"),
		PublicURL:   form.Get("public_url"),
		ProviderRef: form.Get("provider_ref"),
		AltText:     form.Get("alt_text"),
	}
	if caption := form.Get("caption"); caption != "" {
		asset.Metadata = map[string]any{"caption": caption}
	}
	if raw := strings.TrimSpace(form.Get("metadata")); raw != "" {
		if asset.Metadata == nil {
			asset.Metadata = map[string]any{}
		}
		if err := json.Unmarshal([]byte(raw), &asset.Metadata); err != nil {
			return media.Asset{}, fmt.Errorf("invalid metadata json")
		}
	}
	return asset, nil
}

func bindSettingValues(form url.Values) []settings.Value {
	values := []settings.Value{}
	seen := map[string]bool{}
	for key, vals := range form {
		if key == "action_token" {
			continue
		}
		settingKey := key
		if strings.HasPrefix(key, "setting:") {
			settingKey = strings.TrimPrefix(key, "setting:")
		}
		if seen[settingKey] {
			continue
		}
		seen[settingKey] = true
		if len(vals) == 0 {
			continue
		}
		values = append(values, settings.Value{Key: settingKey, Value: vals[0], Public: true})
	}
	return values
}

func bindMenu(form url.Values, id string) (menus.Menu, error) {
	if id == "" {
		id = strings.TrimSpace(form.Get("id"))
	}
	if id == "" {
		return menus.Menu{}, fmt.Errorf("menu id is required")
	}
	location := form.Get("location")
	if location == "" {
		return menus.Menu{}, fmt.Errorf("menu location is required")
	}
	menu := menus.Menu{ID: id, Location: location}
	if raw := strings.TrimSpace(form.Get("items")); raw != "" {
		var items []menus.Item
		if err := json.Unmarshal([]byte(raw), &items); err != nil {
			return menus.Menu{}, fmt.Errorf("invalid items json")
		}
		menu.Items = items
	}
	return menu, nil
}

func bindContentType(form url.Values) (contenttype.Type, error) {
	id := contenttype.ID(strings.TrimSpace(form.Get("id")))
	if id == "" {
		return contenttype.Type{}, fmt.Errorf("content type id is required")
	}
	label := form.Get("label")
	if label == "" {
		return contenttype.Type{}, fmt.Errorf("content type label is required")
	}
	item := contenttype.Type{
		ID:               id,
		Label:            label,
		PermalinkPattern: defaultIfBlank(form.Get("permalink_pattern"), "/"+string(id)+"s/{slug}"),
		Public:           true,
		Supports:         contenttype.Supports{Title: true, Editor: true, Excerpt: true, FeaturedMedia: true},
	}
	return item, nil
}

func defaultIfBlank(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
