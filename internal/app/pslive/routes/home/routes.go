// Package home contains the route handlers related to the app's home screen.
package home

import (
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

type Handlers struct {
	r godest.TemplateRenderer
}

func New(r godest.TemplateRenderer) *Handlers {
	return &Handlers{
		r: r,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, ss session.Store) {
	ar := auth.NewHTTPRouter(er, ss)
	ar.GET("/", h.HandleHomeGet())
}

func (h *Handlers) HandleHomeGet() auth.HTTPHandlerFunc {
	t := "home/home.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, a)
	}
}
