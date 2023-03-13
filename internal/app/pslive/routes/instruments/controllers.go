package instruments

import (
	"context"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func (h *Handlers) HandleInstrumentControllersPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, iid instruments.InstrumentID, enabled bool, url, protocol string,
		) error {
			controllerID, err := h.is.AddController(ctx, instruments.Controller{
				InstrumentID: iid,
				Enabled:      enabled,
				Protocol:     protocol,
				URL:          url,
			})
			if err != nil {
				return err
			}
			if !enabled {
				return nil
			}
			return h.pco.Add(planktoscope.ClientID(controllerID), url)
		},
	)
}
