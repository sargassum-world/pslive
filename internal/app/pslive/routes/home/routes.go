// Package home contains the route handlers related to the app's home screen.
package home

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Handlers struct {
	r godest.TemplateRenderer

	oc *ory.Client

	is *instruments.Store
	ps *presence.Store
}

func New(
	r godest.TemplateRenderer, oc *ory.Client, is *instruments.Store, ps *presence.Store,
) *Handlers {
	return &Handlers{
		r:  r,
		oc: oc,
		is: is,
		ps: ps,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, ss *session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	hr.GET("/", h.HandleHomeGet())
}

type HomeViewData struct {
	CameraInstruments []instruments.Instrument
	AdminIdentifiers  map[string]string
	PresenceCounts    map[int64]int
}

func getHomeViewData(
	ctx context.Context, oc *ory.Client, is *instruments.Store, ps *presence.Store,
) (vd HomeViewData, err error) {
	in, err := is.GetInstruments(ctx)
	if err != nil {
		return HomeViewData{}, err
	}
	vd.CameraInstruments = make([]instruments.Instrument, 0, len(in))
	for _, instrument := range in {
		if len(instrument.Cameras) > 0 {
			vd.CameraInstruments = append(vd.CameraInstruments, instrument)
		}
	}

	vd.AdminIdentifiers = make(map[string]string)
	for _, instrument := range vd.CameraInstruments {
		if vd.AdminIdentifiers[instrument.AdminID], err = oc.GetIdentifier(
			ctx, instrument.AdminID,
		); err != nil {
			// TODO: log the error
			continue
		}
	}

	vd.PresenceCounts = make(map[int64]int)
	for _, instrument := range vd.CameraInstruments {
		topic := fmt.Sprintf("/instruments/%d/users", instrument.ID)
		vd.PresenceCounts[instrument.ID] = ps.Count(topic)
	}

	return vd, err
}

func (h *Handlers) HandleHomeGet() auth.HTTPHandlerFunc {
	t := "home/home.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		ctx := c.Request().Context()
		homeViewData, err := getHomeViewData(ctx, h.oc, h.is, h.ps)
		if err != nil {
			return err
		}
		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, homeViewData, a)
	}
}
