// Package web contains web application-specific components: static web assets,
// server-side templates, and modest JS sprinkles and spots.
package web

import (
	"embed"
	"html/template"
	"io/fs"

	"github.com/benbjohnson/hashfs"

	"github.com/sargassum-world/fluitans/pkg/godest"
)

// Embeds are embedded filesystems

var (
	//go:embed static/*
	staticEFS   embed.FS
	staticFS, _ = fs.Sub(staticEFS, "static")
	staticHFS   = hashfs.NewFS(staticFS)
)

var (
	//go:embed templates/*
	templatesEFS   embed.FS
	templatesFS, _ = fs.Sub(templatesEFS, "templates")
)

var (
	//go:embed app/public/build/*
	appEFS   embed.FS
	appFS, _ = fs.Sub(appEFS, "app/public/build")
	appHFS   = hashfs.NewFS(appFS)
)

var (
	//go:embed app/public/build/fonts/*
	fontsEFS   embed.FS
	fontsFS, _ = fs.Sub(fontsEFS, "app/public/build/fonts")
)

//go:embed app/public/build/bundle-eager.js
var bundleEagerJS string

//go:embed app/public/build/theme-eager.min.css
var bundleEagerCSS string

func NewEmbeds() godest.Embeds {
	return godest.Embeds{
		StaticFS:    staticFS,
		StaticHFS:   staticHFS,
		TemplatesFS: templatesFS,
		AppFS:       appFS,
		AppHFS:      appHFS,
		FontsFS:     fontsFS,
	}
}

// Inlines are strings to include in-line in templates

func NewInlines() godest.Inlines {
	return godest.Inlines{
		CSS: map[string]template.CSS{
			"BundleEager": template.CSS(bundleEagerCSS),
		},
		JS: map[string]template.JS{
			//nolint:gosec // This is generated from code in web/app/src, so we know it's well-formed
			"BundleEager": template.JS(bundleEagerJS),
		},
	}
}
