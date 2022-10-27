package handling

import (
	"time"

	"github.com/gorilla/sessions"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

func replacePresenceListStream(topic, partial string, ps *presence.Store) turbostreams.Message {
	known, anonymous := ps.List(topic)
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   topic + "/list",
		Template: partial,
		Data: map[string]interface{}{
			"Topic":     topic,
			"Known":     known,
			"Anonymous": anonymous,
		},
	}
}

func replacePresenceCountStream(topic, partial string, ps *presence.Store) turbostreams.Message {
	count := ps.Count(topic)
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   topic + "/count",
		Template: partial,
		Data: map[string]interface{}{
			"Topic": topic,
			"Count": count,
		},
	}
}

const (
	usersListPartial  = "shared/presence/users-list.partial.tmpl"
	usersCountPartial = "shared/presence/users-count.partial.tmpl"
	subPubDelay       = 100 // ms; delay the pub so that we can update the page whose GET caused the sub
)

func HandlePresenceSub(
	r godest.TemplateRenderer, ss *session.Store, oc *ory.Client, ps *presence.Store,
) turbostreams.HandlerFunc {
	tList := usersListPartial
	r.MustHave(tList)
	tCount := usersCountPartial
	r.MustHave(tCount)
	return auth.HandleTSWithSession(
		func(c *turbostreams.Context, a auth.Auth, sess *sessions.Session) (err error) {
			if a.Identity.User != "" && !ps.IsKnown(sess.ID) {
				user, err := oc.GetIdentifier(c.Context(), a.Identity.User)
				if err != nil {
					return err
				}
				ps.Remember(sess.ID, a.Identity.User, user)
			}
			if ps.Add(c.Topic(), sess.ID) {
				go func() {
					time.Sleep(subPubDelay * time.Millisecond)
					c.Broadcast(c.Topic()+"/list", replacePresenceListStream(c.Topic(), tList, ps))
					c.Broadcast(c.Topic()+"/count", replacePresenceCountStream(c.Topic(), tCount, ps))
				}()
			}
			return nil
		},
		ss,
	)
}

func HandlePresenceUnsub(
	r godest.TemplateRenderer, ss *session.Store, ps *presence.Store,
) turbostreams.HandlerFunc {
	tList := usersListPartial
	r.MustHave(tList)
	tCount := usersCountPartial
	r.MustHave(tCount)
	return auth.HandleTSWithSession(
		func(c *turbostreams.Context, a auth.Auth, sess *sessions.Session) error {
			if ps.Remove(c.Topic(), sess.ID) {
				c.Broadcast(c.Topic()+"/list", replacePresenceListStream(c.Topic(), tList, ps))
				c.Broadcast(c.Topic()+"/count", replacePresenceCountStream(c.Topic(), tCount, ps))
			}
			return nil
		},
		ss,
	)
}
