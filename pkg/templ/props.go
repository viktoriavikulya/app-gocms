package views

import "github.com/fastygo/platform/pkg/render"

// PageProps carries admin shell data for the product page layout.
type PageProps struct {
	Title  string
	Screen render.ScreenModel
	Nav    []NavLinkData
}

// NavLinkData is one sidebar navigation entry.
type NavLinkData struct {
	Href  string
	Label string
}
