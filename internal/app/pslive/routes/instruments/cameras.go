package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentCamerasPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, id instruments.InstrumentID, enabled bool, params url.Values,
		) error {
			protocol := params.Get("protocol")
			url := params.Get("url")
			_, err := h.is.AddCamera(ctx, instruments.Camera{
				InstrumentID: id,
				Enabled:      enabled,
				Protocol:     protocol,
				URL:          url,
			})
			return err
		},
	)
}
