// Package routes contains the route handlers for the Planktoscope Live server.
package routes

import (
	"github.com/sargassum-world/fluitans/pkg/godest"

	"github.com/sargassum-world/pslive/internal/app/pslive/client"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/assets"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/home"
)

type Handlers struct {
	r       godest.TemplateRenderer
	clients *client.Clients
}

func New(r godest.TemplateRenderer, clients *client.Clients) *Handlers {
	return &Handlers{
		r:       r,
		clients: clients,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, em godest.Embeds) {
	assets.RegisterStatic(er, em)
	assets.NewTemplated(h.r).Register(er)
	home.New(h.r, h.clients.Sessions).Register(er)
	auth.New(h.r, h.clients.Authn, h.clients.Sessions).Register(er)
}
