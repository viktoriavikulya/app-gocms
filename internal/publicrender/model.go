package publicrender

import (
	"time"

	"github.com/fastygo/app-gocms/internal/domain/content"
	"github.com/fastygo/app-gocms/internal/domain/menus"
)

type Kind string

const (
	KindHome     Kind = "home"
	KindPost     Kind = "post"
	KindPage     Kind = "page"
	KindNotFound Kind = "not_found"
)

type Site struct {
	Title       string
	Description string
	Locale      string
}

type MenuItem struct {
	Label string
	URL   string
}

type Page struct {
	Kind         Kind
	TemplateRole string
	StatusCode   int
	Site         Site
	Title        string
	Slug         string
	Permalink    string
	Content      string
	Excerpt      string
	PublishedAt  time.Time
	Entries      []Entry
	Menu         []MenuItem
	ThemeID      string
}

type Entry struct {
	Title       string
	Slug        string
	Permalink   string
	Excerpt     string
	PublishedAt time.Time
}

func MenuFromDomain(menu menus.Menu) []MenuItem {
	items := []MenuItem{}
	for _, item := range menu.Items {
		items = append(items, MenuItem{Label: item.Label, URL: item.URL})
	}
	return items
}

func EntryFromContent(entry content.Entry, permalink string) Entry {
	return Entry{
		Title:       Title(entry),
		Slug:        entry.Slug,
		Permalink:   permalink,
		Excerpt:     entry.Excerpt,
		PublishedAt: entry.PublishedAt,
	}
}

func Title(entry content.Entry) string {
	if entry.Title == nil {
		return string(entry.ID)
	}
	if value := entry.Title["en"]; value != "" {
		return value
	}
	for _, value := range entry.Title {
		if value != "" {
			return value
		}
	}
	return string(entry.ID)
}
