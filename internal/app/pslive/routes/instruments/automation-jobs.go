package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentAutomationJobsPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(
			ctx context.Context, iid instruments.InstrumentID,
			enabled bool, name, description string, params url.Values,
		) error {
			specType := params.Get("type")
			specification := params.Get("specification")
			id, err := h.is.AddAutomationJob(ctx, instruments.AutomationJob{
				InstrumentID:  iid,
				Enabled:       enabled,
				Name:          name,
				Description:   description,
				Type:          specType,
				Specification: specification,
			})
			if err != nil {
				return err
			}
			if !enabled {
				return nil
			}
			return h.ijo.Add(id, iid, name, specType, specification)
		},
	)
}
