// Package handling provides utilities for handlers
package handling

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

// Rendering

func AddAuthData(a auth.Auth, messages []turbostreams.Message) ([]turbostreams.Message, error) {
	published := make([]turbostreams.Message, len(messages))
	for i, m := range messages {
		d, ok := m.Data.(map[string]interface{})
		if !ok {
			return nil, errors.Errorf("unexpected turbo stream message data type: %T", m.Data)
		}

		data := make(map[string]interface{})
		for key, value := range d {
			data[key] = value
		}
		data["Auth"] = a
		published[i] = turbostreams.Message{
			Action:   m.Action,
			Target:   m.Target,
			Template: m.Template,
			Data:     data,
		}
	}
	return published, nil
}

func HandleTSMsg(r godest.TemplateRenderer, ss session.Store) turbostreams.HandlerFunc {
	return auth.HandleTS(
		func(c turbostreams.Context, a auth.Auth) (err error) {
			// Render with auth data
			published, err := AddAuthData(a, c.Published())
			if err != nil {
				return err
			}
			return r.WriteTurboStream(c.MsgWriter(), published...)
		},
		ss,
	)
}
