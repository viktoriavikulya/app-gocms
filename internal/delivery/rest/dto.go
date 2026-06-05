// Package rest exposes GoCMS-compatible REST JSON projections.
// AppCMS currently emits a single locale ("en") until i18n lands (Phase 11).
package rest

import (
	"context"
	"strings"
	"time"

	domaincontent "github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	domainrevisions "github.com/fastygo/app-gocms/internal/domain/revisions"
	"github.com/fastygo/app-gocms/internal/domain/media"
	"github.com/fastygo/app-gocms/internal/domain/menus"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/domain/taxonomy"
	domainusers "github.com/fastygo/app-gocms/internal/domain/users"
)

const defaultLocale = "en"

type ContentDTO struct {
	ID              string            `json:"id"`
	Kind            string            `json:"kind"`
	Status          string            `json:"status"`
	Slug            map[string]string `json:"slug"`
	Title           map[string]string `json:"title"`
	Content         map[string]string `json:"content"`
	Excerpt         map[string]string `json:"excerpt"`
	AuthorID        string            `json:"author_id"`
	FeaturedMediaID string            `json:"featured_media_id,omitempty"`
	TaxonomyIDs     []string          `json:"taxonomy_ids"`
	Metadata        map[string]any    `json:"metadata,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	PublishedAt     *time.Time        `json:"published_at,omitempty"`
	Links           map[string]string `json:"links"`
}

type ContentTypeDTO struct {
	ID             string               `json:"id"`
	Label          string               `json:"label"`
	Public         bool                 `json:"public"`
	RESTVisible    bool                 `json:"rest_visible"`
	GraphQLVisible bool                 `json:"graphql_visible"`
	Supports       contenttype.Supports `json:"supports"`
	Archive        bool                 `json:"archive"`
	Permalink      string               `json:"permalink"`
}

type TaxonomyDTO struct {
	Type            string   `json:"type"`
	Label           string   `json:"label"`
	Mode            string   `json:"mode"`
	AssignedToKinds []string `json:"assigned_to_kinds"`
	Public          bool     `json:"public"`
	RESTVisible     bool     `json:"rest_visible"`
	GraphQLVisible  bool     `json:"graphql_visible"`
}

type TermDTO struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Name        map[string]string `json:"name"`
	Slug        map[string]string `json:"slug"`
	Description map[string]string `json:"description"`
	ParentID    string            `json:"parent_id,omitempty"`
}

type MediaDTO struct {
	ID        string         `json:"id"`
	Filename  string         `json:"filename"`
	MimeType  string         `json:"mime_type"`
	SizeBytes int64          `json:"size_bytes"`
	Width     int            `json:"width,omitempty"`
	Height    int            `json:"height,omitempty"`
	AltText   string         `json:"alt_text,omitempty"`
	Caption   string         `json:"caption,omitempty"`
	PublicURL string         `json:"public_url"`
	Provider  string         `json:"provider,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Variants  map[string]any `json:"variants,omitempty"`
}

type AuthorDTO struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	DisplayName   string `json:"display_name"`
	Bio           string `json:"bio,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	WebsiteURL    string `json:"website_url,omitempty"`
}

type SettingDTO struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

type MenuDTO struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Location string        `json:"location"`
	Items    []MenuItemDTO `json:"items"`
}

type MenuItemDTO struct {
	ID       string        `json:"id"`
	Label    string        `json:"label"`
	URL      string        `json:"url"`
	Kind     string        `json:"kind,omitempty"`
	TargetID string        `json:"target_id,omitempty"`
	Children []MenuItemDTO `json:"children,omitempty"`
}

type RevisionDTO struct {
	ID        string     `json:"id"`
	EntryID   string     `json:"entry_id"`
	Snapshot  ContentDTO `json:"snapshot"`
	CreatedAt time.Time  `json:"created_at"`
}

type MediaResolver interface {
	Get(ctx context.Context, id string) (media.Asset, bool, error)
}

func localize(value string) map[string]string {
	if value == "" {
		return map[string]string{}
	}
	return map[string]string{defaultLocale: value}
}

func delocalize(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	if value, ok := values[defaultLocale]; ok {
		return value
	}
	for _, value := range values {
		return value
	}
	return ""
}

func IsPublicEntry(entry domaincontent.Entry, now time.Time) bool {
	return entry.IsPublicAt(now)
}

func ContentProjection(entry domaincontent.Entry, includePrivate bool) ContentDTO {
	metadata := filterMetadata(entry.Metadata, includePrivate)
	taxonomyIDs := append([]string{}, entry.TermIDs...)
	var publishedAt *time.Time
	if !entry.PublishedAt.IsZero() {
		publishedAt = &entry.PublishedAt
	}
	collection := string(entry.Kind) + "s"
	if entry.Kind == domaincontent.KindPage {
		collection = "pages"
	}
	return ContentDTO{
		ID:              string(entry.ID),
		Kind:            string(entry.Kind),
		Status:          string(entry.Status),
		Slug:            localize(entry.Slug),
		Title:           cloneLocalized(entry.Title),
		Content:         localize(entry.Content),
		Excerpt:         localize(entry.Excerpt),
		AuthorID:        entry.AuthorID,
		FeaturedMediaID: entry.FeaturedMediaID,
		TaxonomyIDs:     taxonomyIDs,
		Metadata:        metadata,
		CreatedAt:       entry.CreatedAt,
		UpdatedAt:       entry.UpdatedAt,
		PublishedAt:     publishedAt,
		Links: map[string]string{
			"self": "/go-json/go/v2/" + collection + "/" + string(entry.ID),
		},
	}
}

func filterMetadata(metadata map[string]any, includePrivate bool) map[string]any {
	if len(metadata) == 0 {
		return nil
	}
	if includePrivate {
		return cloneMap(metadata)
	}
	result := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if strings.HasSuffix(key, "_private") {
			continue
		}
		result[key] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func ContentTypeProjection(item contenttype.Type) ContentTypeDTO {
	return ContentTypeDTO{
		ID:             string(item.ID),
		Label:          item.Label,
		Public:         item.Public,
		RESTVisible:    item.Public,
		GraphQLVisible: item.Public,
		Supports:       item.Supports,
		Archive:        false,
		Permalink:      item.PermalinkPattern,
	}
}

func TaxonomyProjection(definition taxonomy.Definition) TaxonomyDTO {
	return TaxonomyDTO{
		Type:            definition.Type,
		Label:           definition.Label,
		Mode:            string(definition.Mode),
		AssignedToKinds: append([]string{}, definition.AssignedKinds...),
		Public:          definition.Public,
		RESTVisible:     definition.Public,
		GraphQLVisible:  definition.Public,
	}
}

func TermProjection(term taxonomy.Term) TermDTO {
	return TermDTO{
		ID:          term.ID,
		Type:        term.TaxonomyType,
		Name:        localize(term.Name),
		Slug:        localize(term.Slug),
		Description: localize(term.Description),
		ParentID:    term.ParentID,
	}
}

func MediaProjection(asset media.Asset) MediaDTO {
	filename := asset.Title
	if filename == "" {
		filename = asset.ID
	}
	caption := ""
	sizeBytes := int64(0)
	width := 0
	height := 0
	if asset.Metadata != nil {
		if value, ok := asset.Metadata["caption"].(string); ok {
			caption = value
		}
		if value, ok := asset.Metadata["size_bytes"].(float64); ok {
			sizeBytes = int64(value)
		}
		if value, ok := asset.Metadata["width"].(float64); ok {
			width = int(value)
		}
		if value, ok := asset.Metadata["height"].(float64); ok {
			height = int(value)
		}
	}
	return MediaDTO{
		ID:        asset.ID,
		Filename:  filename,
		MimeType:  asset.MIMEType,
		SizeBytes: sizeBytes,
		Width:     width,
		Height:    height,
		AltText:   asset.AltText,
		Caption:   caption,
		PublicURL: asset.PublicURL,
		Provider:  asset.ProviderRef,
		Metadata:  cloneMap(asset.Metadata),
		Variants:  cloneMap(asset.Variants),
	}
}

func AuthorProjection(author domainusers.AuthorProfile, avatarURL string) AuthorDTO {
	return AuthorDTO{
		ID:            author.ID,
		Slug:          author.Slug,
		DisplayName:   author.DisplayName,
		Bio:           author.Bio,
		AvatarURL:     avatarURL,
		AvatarMediaID: author.AvatarID,
	}
}

func SettingProjection(value settings.Value) SettingDTO {
	return SettingDTO{Key: value.Key, Value: value.Value}
}

func MenuProjection(menu menus.Menu) MenuDTO {
	name := menu.Location
	if name == "" {
		name = menu.ID
	}
	return MenuDTO{
		ID:       menu.ID,
		Name:     name,
		Location: menu.Location,
		Items:    menuItemsTree(menu.Items),
	}
}

func menuItemsTree(items []menus.Item) []MenuItemDTO {
	children := make(map[string][]menus.Item)
	roots := []menus.Item{}
	for _, item := range items {
		if item.ParentID == "" {
			roots = append(roots, item)
			continue
		}
		children[item.ParentID] = append(children[item.ParentID], item)
	}
	result := make([]MenuItemDTO, 0, len(roots))
	for _, item := range roots {
		result = append(result, menuItemDTO(item, children))
	}
	return result
}

func menuItemDTO(item menus.Item, children map[string][]menus.Item) MenuItemDTO {
	dto := MenuItemDTO{
		ID:    item.ID,
		Label: item.Label,
		URL:   item.URL,
	}
	for _, child := range children[item.ID] {
		dto.Children = append(dto.Children, menuItemDTO(child, children))
	}
	return dto
}

func cloneLocalized(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func cloneMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func projectContentList(entries []domaincontent.Entry, includePrivate bool) []ContentDTO {
	result := make([]ContentDTO, 0, len(entries))
	for _, entry := range entries {
		result = append(result, ContentProjection(entry, includePrivate))
	}
	return result
}

func projectContentTypes(items []contenttype.Type) []ContentTypeDTO {
	result := make([]ContentTypeDTO, 0, len(items))
	for _, item := range items {
		result = append(result, ContentTypeProjection(item))
	}
	return result
}

func projectTaxonomies(items []taxonomy.Definition) []TaxonomyDTO {
	result := make([]TaxonomyDTO, 0, len(items))
	for _, item := range items {
		result = append(result, TaxonomyProjection(item))
	}
	return result
}

func projectTerms(items []taxonomy.Term) []TermDTO {
	result := make([]TermDTO, 0, len(items))
	for _, item := range items {
		result = append(result, TermProjection(item))
	}
	return result
}

func projectMediaList(items []media.Asset) []MediaDTO {
	result := make([]MediaDTO, 0, len(items))
	for _, item := range items {
		result = append(result, MediaProjection(item))
	}
	return result
}

func projectSettings(items []settings.Value) []SettingDTO {
	result := make([]SettingDTO, 0, len(items))
	for _, item := range items {
		result = append(result, SettingProjection(item))
	}
	return result
}

func projectPublicSettings(items []settings.Value) []SettingDTO {
	result := make([]SettingDTO, 0, len(items))
	for _, item := range items {
		if item.Private || strings.HasSuffix(item.Key, "_private") {
			continue
		}
		result = append(result, SettingProjection(item))
	}
	return result
}

func projectMenus(items []menus.Menu) []MenuDTO {
	result := make([]MenuDTO, 0, len(items))
	for _, item := range items {
		result = append(result, MenuProjection(item))
	}
	return result
}

func resolveAvatarURL(ctx context.Context, resolver MediaResolver, avatarID string) string {
	if avatarID == "" || resolver == nil {
		return ""
	}
	asset, ok, err := resolver.Get(ctx, avatarID)
	if err != nil || !ok {
		return ""
	}
	return asset.PublicURL
}

func RevisionProjection(revision domainrevisions.Revision, includePrivate bool) RevisionDTO {
	return RevisionDTO{
		ID:        revision.ID,
		EntryID:   string(revision.EntryID),
		Snapshot:  ContentProjection(revision.Entry, includePrivate),
		CreatedAt: revision.CreatedAt,
	}
}

func projectRevisions(items []domainrevisions.Revision, includePrivate bool) []RevisionDTO {
	result := make([]RevisionDTO, 0, len(items))
	for _, item := range items {
		result = append(result, RevisionProjection(item, includePrivate))
	}
	return result
}
