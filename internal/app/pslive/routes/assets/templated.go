// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"github.com/labstack/echo/v4"
)

func (h *TemplatedHandlers) getWebmanifest() echo.HandlerFunc {
	t := "app/app.webmanifest.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context) error {
		// Produce output
		c.Response().Header().Set(echo.HeaderContentType, "application/manifest+json")
		return h.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, struct{}{})
	}
}
