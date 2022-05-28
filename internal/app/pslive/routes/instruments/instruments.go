package instruments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atrox/haikunatorgo"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type InstrumentsViewData struct {
	Instruments      []instruments.Instrument
	AdminIdentifiers map[string]string
}

func getInstrumentsViewData(
	ctx context.Context, oc *ory.Client, is *instruments.Store,
) (vd InstrumentsViewData, err error) {
	if vd.Instruments, err = is.GetInstruments(ctx); err != nil {
		return InstrumentsViewData{}, err
	}

	vd.AdminIdentifiers = make(map[string]string)
	for _, instrument := range vd.Instruments {
		if vd.AdminIdentifiers[instrument.AdminID], err = oc.GetIdentifier(
			ctx, instrument.AdminID,
		); err != nil {
			// TODO: log the error
			continue
		}
	}

	return vd, err
}

func (h *Handlers) HandleInstrumentsGet() auth.HTTPHandlerFunc {
	t := "instruments/instruments.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		instrumentsViewData, err := getInstrumentsViewData(c.Request().Context(), h.oc, h.is)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, instrumentsViewData, a)
	}
}

func (h *Handlers) HandleInstrumentsPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		i := instruments.Instrument{
			Name:        haikunator.New().Haikunate(),
			Description: "An unknown instrument!",
			AdminID:     a.Identity.User,
		}
		id, err := h.is.AddInstrument(c.Request().Context(), i)
		if err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}
