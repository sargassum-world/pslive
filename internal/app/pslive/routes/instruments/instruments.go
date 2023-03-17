package instruments

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atrox/haikunatorgo"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type InstrumentsViewData struct {
	Instruments      []instruments.Instrument
	AdminIdentifiers map[instruments.AdminID]ory.IdentityIdentifier
}

func getInstrumentsViewData(
	ctx context.Context, oc *ory.Client, is *instruments.Store,
) (vd InstrumentsViewData, err error) {
	if vd.Instruments, err = is.GetInstruments(ctx); err != nil {
		return InstrumentsViewData{}, err
	}

	vd.AdminIdentifiers = make(map[instruments.AdminID]ory.IdentityIdentifier)
	for _, instrument := range vd.Instruments {
		if vd.AdminIdentifiers[instrument.AdminID], err = oc.GetIdentifier(
			ctx, ory.IdentityID(instrument.AdminID),
		); err != nil {
			// TODO: log the error
			continue
		}
	}

	return vd, err
}

type InstrumentsViewAuthz struct {
	CreateInstrument bool
}

func getInstrumentsViewAuthz(
	ctx context.Context, a auth.Auth, azc *auth.AuthzChecker,
) (authz InstrumentsViewAuthz, err error) {
	path := "/instruments"
	if authz.CreateInstrument, err = azc.Allow(ctx, a, path, http.MethodPost, nil); err != nil {
		return InstrumentsViewAuthz{}, errors.Wrap(err, "couldn't check authz for creating instrument")
	}
	return authz, nil
}

func (h *Handlers) HandleInstrumentsGet() auth.HTTPHandlerFunc {
	t := "instruments/instruments.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		ctx := c.Request().Context()
		instrumentsViewData, err := getInstrumentsViewData(ctx, h.oc, h.is)
		if err != nil {
			return err
		}
		if a.Authorizations, err = getInstrumentsViewAuthz(ctx, a, h.azc); err != nil {
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
			AdminID:     instruments.AdminID(a.Identity.User),
		}
		id, err := h.is.AddInstrument(c.Request().Context(), i)
		if err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", int64(id)))
	}
}
