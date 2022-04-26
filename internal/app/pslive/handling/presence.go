package handling

import (
	"time"

	"github.com/gorilla/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

func replacePresenceStream(topic, partial string, ps *presence.Store) turbostreams.Message {
	known, anonymous := ps.List(topic)
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   topic,
		Template: partial,
		Data: map[string]interface{}{
			"Topic":     topic,
			"Known":     known,
			"Anonymous": anonymous,
		},
	}
}

const (
	usersPartial = "shared/presence/users.partial.tmpl"
	subPubDelay  = 100 // ms; delay the pub so that we can update the page whose GET caused the sub
)

func HandlePresenceSub(
	r godest.TemplateRenderer, ss session.Store, oc *ory.Client, ps *presence.Store,
) turbostreams.HandlerFunc {
	t := usersPartial
	r.MustHave(t)
	return auth.HandleTSWithSession(
		func(c turbostreams.Context, a auth.Auth, sess *sessions.Session) (err error) {
			if a.Identity.User != "" && !ps.IsKnown(sess.ID) {
				user, err := oc.GetIdentifier(c.Context(), a.Identity.User)
				if err != nil {
					return err
				}
				ps.Remember(sess.ID, a.Identity.User, user)
			}
			ps.Add(c.Topic(), sess.ID)
			go func() {
				time.Sleep(subPubDelay * time.Millisecond)
				c.Publish(replacePresenceStream(c.Topic(), t, ps))
			}()
			return nil
		},
		ss,
	)
}

func HandlePresenceUnsub(
	r godest.TemplateRenderer, ss session.Store, ps *presence.Store,
) turbostreams.HandlerFunc {
	t := usersPartial
	r.MustHave(t)
	return auth.HandleTSWithSession(
		func(c turbostreams.Context, a auth.Auth, sess *sessions.Session) error {
			ps.Remove(c.Topic(), sess.ID)
			c.Publish(replacePresenceStream(c.Topic(), t, ps))
			return nil
		},
		ss,
	)
}
