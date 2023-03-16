package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

func (h *Handlers) HandleInstrumentControllersPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, iid instruments.InstrumentID,
			enabled bool, name, description string, params url.Values,
		) error {
			protocol := params.Get("protocol")
			url := params.Get("url")
			controllerID, err := h.is.AddController(ctx, instruments.Controller{
				InstrumentID: iid,
				Enabled:      enabled,
				Name:         name,
				Description:  description,
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
