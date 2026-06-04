package publicsite

import (
	"context"
	"net/http"
	"strings"

	appcontent "github.com/fastygo/app-gocms/internal/application/content"
	appcontenttype "github.com/fastygo/app-gocms/internal/application/contenttype"
	appmenus "github.com/fastygo/app-gocms/internal/application/menus"
	appsettings "github.com/fastygo/app-gocms/internal/application/settings"
	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/contenttype"
	"github.com/fastygo/app-gocms/internal/domain/settings"
	"github.com/fastygo/app-gocms/internal/permalinks"
	"github.com/fastygo/app-gocms/internal/publicrender"
)

type services struct {
	content      appcontent.Service
	contentTypes appcontenttype.Service
	settings     appsettings.Service
	menus        appmenus.Service
}

type assembler struct {
	services services
}

func newAssembler(services services) assembler {
	return assembler{services: services}
}

func (a assembler) Assemble(ctx context.Context, path string) (publicrender.Page, error) {
	site, themeID, err := a.site(ctx)
	if err != nil {
		return publicrender.Page{}, err
	}
	settings, err := a.permalinkSettings(ctx)
	if err != nil {
		return publicrender.Page{}, err
	}
	menu := a.menu(ctx)
	for _, candidate := range permalinks.Resolve(path, settings) {
		switch candidate.Kind {
		case permalinks.CandidateHome:
			posts, err := a.publishedEntries(ctx, content.KindPost)
			if err != nil {
				return publicrender.Page{}, err
			}
			return publicrender.Page{Kind: publicrender.KindHome, TemplateRole: "front", StatusCode: http.StatusOK, Site: site, Title: site.Title, Entries: posts, Menu: menu, ThemeID: themeID}, nil
		case permalinks.CandidatePage:
			if page, ok, err := a.entryBySlug(ctx, content.KindPage, candidate.Slug, settings, site, menu, themeID); err != nil || ok {
				return page, err
			}
		case permalinks.CandidatePost:
			if page, ok, err := a.entryBySlug(ctx, content.KindPost, candidate.Slug, settings, site, menu, themeID); err != nil || ok {
				return page, err
			}
		}
	}
	return publicrender.Page{Kind: publicrender.KindNotFound, TemplateRole: "not_found", StatusCode: http.StatusNotFound, Site: site, Title: "Not found", Menu: menu, ThemeID: themeID}, nil
}

func (a assembler) entryBySlug(ctx context.Context, kind content.Kind, slug string, settings permalinks.Settings, site publicrender.Site, menu []publicrender.MenuItem, themeID string) (publicrender.Page, bool, error) {
	items, err := a.publishedContent(ctx, kind)
	if err != nil {
		return publicrender.Page{}, false, err
	}
	for _, item := range items {
		if item.Slug != slug {
			continue
		}
		pageKind := publicrender.KindPost
		role := "post"
		if kind == content.KindPage {
			pageKind = publicrender.KindPage
			role = "page"
		}
		return publicrender.Page{
			Kind:         pageKind,
			TemplateRole: role,
			StatusCode:   http.StatusOK,
			Site:         site,
			Title:        publicrender.Title(item),
			Slug:         item.Slug,
			Permalink:    permalinks.EntryPath(item, settings),
			Content:      item.Content,
			Excerpt:      item.Excerpt,
			PublishedAt:  item.PublishedAt,
			Menu:         menu,
			ThemeID:      themeID,
		}, true, nil
	}
	return publicrender.Page{}, false, nil
}

func (a assembler) publishedEntries(ctx context.Context, kind content.Kind) ([]publicrender.Entry, error) {
	settings, err := a.permalinkSettings(ctx)
	if err != nil {
		return nil, err
	}
	items, err := a.publishedContent(ctx, kind)
	if err != nil {
		return nil, err
	}
	entries := []publicrender.Entry{}
	for _, item := range items {
		entries = append(entries, publicrender.EntryFromContent(item, permalinks.EntryPath(item, settings)))
	}
	return entries, nil
}

func (a assembler) publishedContent(ctx context.Context, kind content.Kind) ([]content.Entry, error) {
	items, err := a.services.content.List(ctx, content.Query{Kind: kind, Status: content.StatusPublished})
	if err != nil {
		return nil, err
	}
	public := []content.Entry{}
	for _, item := range items {
		if item.Visibility == content.VisibilityPublic {
			public = append(public, item)
		}
	}
	return public, nil
}

func (a assembler) site(ctx context.Context) (publicrender.Site, string, error) {
	site := publicrender.Site{Title: "AppCMS", Locale: "en"}
	themeID := "gocms-default"
	values, err := a.services.settings.Public(ctx)
	if err != nil {
		return site, themeID, err
	}
	for _, value := range values {
		switch value.Key {
		case "site.title":
			site.Title = stringValue(value.Value, site.Title)
		case "site.description":
			site.Description = stringValue(value.Value, site.Description)
		case "site.locale":
			site.Locale = stringValue(value.Value, site.Locale)
		case "theme.active":
			themeID = stringValue(value.Value, themeID)
		}
	}
	return site, themeID, nil
}

func (a assembler) permalinkSettings(ctx context.Context) (permalinks.Settings, error) {
	settings := permalinks.DefaultSettings()
	for _, id := range []contenttype.ID{contenttype.Post, contenttype.Page} {
		item, ok, err := a.services.contentTypes.Get(ctx, id)
		if err != nil {
			return settings, err
		}
		if !ok || strings.TrimSpace(item.PermalinkPattern) == "" {
			continue
		}
		if id == contenttype.Post {
			settings.PostPattern = item.PermalinkPattern
		} else {
			settings.PagePattern = item.PermalinkPattern
		}
	}
	return settings, nil
}

func (a assembler) menu(ctx context.Context) []publicrender.MenuItem {
	menu, ok, err := a.services.menus.ByLocation(ctx, "primary")
	if err != nil || !ok {
		return nil
	}
	return publicrender.MenuFromDomain(menu)
}

func stringValue(value any, fallback string) string {
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	return fallback
}

func defaultSettingsRegistry() appsettings.Registry {
	return appsettings.NewRegistry(
		settings.Definition{Key: "site.title", Group: "site", DefaultValue: "AppCMS", Public: true},
		settings.Definition{Key: "site.description", Group: "site", DefaultValue: "A FastyGo powered site.", Public: true},
		settings.Definition{Key: "site.locale", Group: "site", DefaultValue: "en", Public: true},
		settings.Definition{Key: "theme.active", Group: "theme", DefaultValue: "gocms-default", Public: true},
	)
}
