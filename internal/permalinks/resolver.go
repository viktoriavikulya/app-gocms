package permalinks

import (
	"strings"

	"github.com/fastygo/app-gocms/internal/domain/content"
)

type CandidateKind string

const (
	CandidateHome CandidateKind = "home"
	CandidatePage CandidateKind = "page"
	CandidatePost CandidateKind = "post"
	CandidateNone CandidateKind = "none"
)

type Candidate struct {
	Kind CandidateKind
	Slug string
}

type Settings struct {
	PostPattern string
	PagePattern string
}

func DefaultSettings() Settings {
	return Settings{PostPattern: "/posts/{slug}", PagePattern: "/{slug}"}
}

func Resolve(path string, settings Settings) []Candidate {
	clean := cleanPath(path)
	if clean == "/" {
		return []Candidate{{Kind: CandidateHome}}
	}
	slug := strings.Trim(clean, "/")
	candidates := []Candidate{}
	if matchesPattern(clean, settings.PagePattern) {
		candidates = append(candidates, Candidate{Kind: CandidatePage, Slug: slugFromPattern(clean, settings.PagePattern)})
	}
	if matchesPattern(clean, settings.PostPattern) {
		candidates = append(candidates, Candidate{Kind: CandidatePost, Slug: slugFromPattern(clean, settings.PostPattern)})
	}
	if len(candidates) == 0 && slug != "" && !strings.Contains(slug, "/") {
		candidates = append(candidates, Candidate{Kind: CandidatePage, Slug: slug}, Candidate{Kind: CandidatePost, Slug: slug})
	}
	return candidates
}

func EntryPath(entry content.Entry, settings Settings) string {
	pattern := settings.PostPattern
	if entry.Kind == content.KindPage {
		pattern = settings.PagePattern
	}
	if strings.TrimSpace(pattern) == "" {
		pattern = "/{slug}"
	}
	path := strings.ReplaceAll(pattern, "{slug}", entry.Slug)
	path = strings.ReplaceAll(path, "%postname%", entry.Slug)
	return cleanPath(path)
}

func cleanPath(path string) string {
	path = "/" + strings.Trim(path, "/")
	if path == "/" {
		return path
	}
	return strings.TrimRight(path, "/")
}

func matchesPattern(path string, pattern string) bool {
	pattern = cleanPath(pattern)
	switch {
	case pattern == "/{slug}":
		return strings.Count(strings.Trim(path, "/"), "/") == 0
	case strings.Contains(pattern, "{slug}"):
		prefix, suffix, _ := strings.Cut(pattern, "{slug}")
		return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) && strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix) != ""
	case strings.Contains(pattern, "%postname%"):
		prefix, suffix, _ := strings.Cut(pattern, "%postname%")
		return strings.HasPrefix(path, prefix) && strings.HasSuffix(path, suffix) && strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix) != ""
	default:
		return path == pattern
	}
}

func slugFromPattern(path string, pattern string) string {
	pattern = cleanPath(pattern)
	token := "{slug}"
	if strings.Contains(pattern, "%postname%") {
		token = "%postname%"
	}
	prefix, suffix, found := strings.Cut(pattern, token)
	if !found {
		return strings.Trim(path, "/")
	}
	return strings.Trim(strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix), "/")
}
