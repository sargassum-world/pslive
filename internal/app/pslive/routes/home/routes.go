// Package home contains the route handlers related to the app's home screen.
package home

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
)

type Handlers struct {
	r  godest.TemplateRenderer
	sc *sessions.Client
}

func New(r godest.TemplateRenderer, sc *sessions.Client) *Handlers {
	return &Handlers{
		r:  r,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	ar := auth.NewAuthAwareRouter(er, h.sc)
	ar.GET("/", h.HandleHomeGet())
}

func (h *Handlers) HandleHomeGet() auth.AuthAwareHandler {
	t := "home/home.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, a)
	}
}
