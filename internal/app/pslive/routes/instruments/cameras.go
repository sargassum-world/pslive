package instruments

import (
	"context"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentCamerasPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, id instruments.InstrumentID,
			enabled bool, protocol, url string,
		) error {
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
