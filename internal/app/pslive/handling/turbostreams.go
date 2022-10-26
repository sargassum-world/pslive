// Package handling provides reusable handlers.
package handling

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

type DataModifier func(
	ctx context.Context, a auth.Auth, data map[string]interface{},
) (modifications map[string]interface{}, err error)

func AddAuthData() DataModifier {
	return func(
		_ context.Context, a auth.Auth, data map[string]interface{},
	) (modifications map[string]interface{}, err error) {
		return map[string]interface{}{
			"Auth": a,
		}, nil
	}
}

func ModifyData(
	ctx context.Context, a auth.Auth, messages []turbostreams.Message, modifiers ...DataModifier,
) ([]turbostreams.Message, error) {
	// TODO: move this function into github.com/sargassum-world/godest/turbostreams?
	// (with a generic type for Auth)
	modified := make([]turbostreams.Message, len(messages))
	for i, m := range messages {
		// Copy the template message
		modified[i] = turbostreams.Message{
			Action:   m.Action,
			Target:   m.Target,
			Template: m.Template,
		}
		if m.Action == turbostreams.ActionRemove {
			// The contents of the stream element will be ignored anyways
			modified[i].Template = ""
			continue
		}

		// Copy the template data
		d, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected turbo stream message data type: %T", m.Data)
		}
		data := make(map[string]interface{})
		for key, value := range d {
			data[key] = value
		}

		// Add data modifications
		for _, modifier := range modifiers {
			modifications, err := modifier(ctx, a, data)
			if err != nil {
				return nil, errors.Wrap(err, "couldn't modify template data")
			}
			for key, value := range modifications {
				data[key] = value
			}
		}

		modified[i].Data = data
	}
	return modified, nil
}

func HandleTSMsg(
	r godest.TemplateRenderer, ss *session.Store, modifiers ...DataModifier,
) turbostreams.HandlerFunc {
	modifiers = append([]DataModifier{AddAuthData()}, modifiers...)
	return auth.HandleTS(
		func(c *turbostreams.Context, a auth.Auth) (err error) {
			// TODO: move this function into github.com/sargassum-world/fluitans/pkg/godest/turbostreams?
			// (without prepending AddAuthData though, and with a generic type for Auth)
			ctx := c.Context()
			modified := c.Published()
			if modified, err = ModifyData(ctx, a, modified, modifiers...); err != nil {
				return err
			}
			return r.WriteTurboStream(c.MsgWriter(), modified...)
		},
		ss,
	)
}
