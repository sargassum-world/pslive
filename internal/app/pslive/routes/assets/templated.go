// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

func (h *TemplatedHandlers) getWebmanifest() echo.HandlerFunc {
	t := "app/app.webmanifest.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context) error {
		const cacheMaxAge = 3600 // 1 hour
		// Produce output
		return h.r.CacheablePage(
			c.Response(), c.Request(), t, struct{}{}, struct{}{},
			godest.WithContentType("application/manifest+json; charset=UTF-8"),
			godest.WithRevalidateWhenStale(cacheMaxAge),
		)
	}
}

func (h *TemplatedHandlers) getOffline() echo.HandlerFunc {
	t := "app/offline.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context) error {
		const cacheMaxAge = 86400 // 1 day
		// Produce output
		return h.r.CacheablePage(
			c.Response(), c.Request(), t, struct{}{}, auth.Auth{},
			godest.WithRevalidateWhenStale(cacheMaxAge),
		)
	}
}
