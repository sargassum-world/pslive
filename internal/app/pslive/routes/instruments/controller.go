package instruments

import (
	"context"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func (h *Handlers) HandleInstrumentControllerPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"controller",
		func(
			ctx context.Context, controllerID instruments.ControllerID, url, protocol string, enabled bool,
		) error {
			if err := h.is.UpdateController(ctx, instruments.Controller{
				ID:       controllerID,
				URL:      url,
				Protocol: protocol,
				Enabled:  enabled,
			}); err != nil {
				return err
			}
			// Note: when we have other controllers, we'll need to generalize this
			if !enabled {
				return h.pco.Remove(ctx, planktoscope.ClientID(controllerID))
			}
			return h.pco.Update(ctx, planktoscope.ClientID(controllerID), url)
		},
		func(ctx context.Context, controllerID instruments.ControllerID) error {
			if err := h.is.DeleteController(ctx, controllerID); err != nil {
				return err
			}
			if err := h.pco.Remove(ctx, planktoscope.ClientID(controllerID)); err != nil {
				return err
			}
			return nil
		},
	)
}
