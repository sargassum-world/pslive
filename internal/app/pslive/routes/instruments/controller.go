package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func (h *Handlers) HandleInstrumentControllerPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"controller",
		func(
			ctx context.Context, id instruments.ControllerID, iid instruments.InstrumentID,
			enabled bool, name, description string, params url.Values,
		) error {
			protocol := params.Get("protocol")
			url := params.Get("url")
			if err := h.is.UpdateController(ctx, instruments.Controller{
				ID:          id,
				Enabled:     enabled,
				Name:        name,
				Description: description,
				Protocol:    protocol,
				URL:         url,
			}); err != nil {
				return err
			}
			// Note: when we have other controllers, we'll need to generalize this
			if !enabled {
				return h.pco.Remove(ctx, planktoscope.ClientID(id))
			}
			return h.pco.Update(ctx, planktoscope.ClientID(id), url)
		},
		func(ctx context.Context, id instruments.ControllerID) error {
			if err := h.is.DeleteController(ctx, id); err != nil {
				return err
			}
			if err := h.pco.Remove(ctx, planktoscope.ClientID(id)); err != nil {
				return err
			}
			return nil
		},
	)
}
