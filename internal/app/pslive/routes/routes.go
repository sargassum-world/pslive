// Package routes contains the route handlers for the Planktoscope Live server.
package routes

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/client"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/assets"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/cable"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/home"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/instruments"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/privatechat"
	"github.com/sargassum-world/pslive/internal/app/pslive/routes/users"
)

type Handlers struct {
	r       godest.TemplateRenderer
	globals *client.Globals
}

func New(r godest.TemplateRenderer, globals *client.Globals) *Handlers {
	return &Handlers{
		r:       r,
		globals: globals,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, em godest.Embeds) {
	ss := h.globals.Sessions
	oc := h.globals.Ory
	azc := h.globals.AuthzChecker
	tsh := h.globals.TSBroker.Hub()
	acc := h.globals.ACCancellers
	l := h.globals.Logger
	is := h.globals.Instruments
	ps := h.globals.Presence
	cs := h.globals.Chat

	assets.RegisterStatic(er, em)
	assets.NewTemplated(h.r).Register(er)
	cable.New(
		h.r, ss, h.globals.CSRFChecker, acc, h.globals.TSSigner, h.globals.TSBroker, l,
	).Register(er)
	home.New(h.r).Register(er, ss)
	auth.New(h.r, ss, oc, acc, ps, l).Register(er)
	instruments.New(
		h.r, oc, azc, tsh, is, h.globals.Planktoscopes, ps, cs,
	).Register(er, tsr, ss)
	privatechat.New(h.r, oc, azc, tsh, ps, cs).Register(er, tsr, ss)
	users.New(h.r, oc, azc, tsh, is, ps, cs).Register(er, tsr, ss)

	tsr.PUB("/*", turbostreams.EmptyHandler)
	tsr.UNSUB("/*", turbostreams.EmptyHandler)
}
