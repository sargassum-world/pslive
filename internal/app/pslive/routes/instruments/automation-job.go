package instruments

import (
	"context"
	"net/url"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
)

func (h *Handlers) HandleInstrumentAutomationJobPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"automationJob",
		func(
			ctx context.Context, id instruments.AutomationJobID, iid instruments.InstrumentID,
			enabled bool, name, description string, params url.Values,
		) error {
			specType := params.Get("type")
			specification := params.Get("specification")
			if err := h.is.UpdateAutomationJob(ctx, instruments.AutomationJob{
				ID:            id,
				Enabled:       enabled,
				Name:          name,
				Description:   description,
				Type:          specType,
				Specification: specification,
			}); err != nil {
				return err
			}
			// Note: when we have other automation job types, we'll need to generalize this
			if !enabled {
				h.ijo.Remove(id)
				return nil
			}
			return h.ijo.Update(id, iid, name, specType, specification)
		},
		func(ctx context.Context, id instruments.AutomationJobID) error {
			if err := h.is.DeleteAutomationJob(ctx, id); err != nil {
				return err
			}
			h.ijo.Remove(id)
			return nil
		},
	)
}
