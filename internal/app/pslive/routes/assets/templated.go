// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest"
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
