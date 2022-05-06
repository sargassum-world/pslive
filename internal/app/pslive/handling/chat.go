package handling

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

const (
	messagePartial = "shared/chat/message.partial.tmpl"
	sendPartial    = "shared/chat/send.partial.tmpl"
)

func appendChatMessageStream(message chat.Message) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionAppend,
		Target:   message.Topic,
		Template: messagePartial,
		Data: map[string]interface{}{
			"Message":          message,
			"AutoscrollOnLoad": true,
		},
	}
}

func replaceChatSendStream(topic string, a auth.Auth) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   topic + "/send",
		Template: sendPartial,
		Data: map[string]interface{}{
			"Topic": topic,
			"Auth":  a,
		},
	}
}

func HandleChatMessagesPost(
	r godest.TemplateRenderer, oc *ory.Client, tsh *turbostreams.MessagesHub, cs *chat.Store,
) auth.HTTPHandlerFunc {
	sendT := sendPartial
	r.MustHave(sendT)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")
		body := c.FormValue("body")
		topic := strings.TrimSuffix(c.Request().URL.Path, "/messages")

		// Run queries
		user, err := oc.GetIdentifier(c.Request().Context(), a.Identity.User)
		if err != nil {
			return err
		}
		// TODO: add a separate MessageViewData type to attach extra data such as sender identifier
		m := chat.Message{
			Topic:            topic + "/messages",
			SendTime:         time.Now(),
			SenderID:         a.Identity.User,
			SenderIdentifier: user,
			Body:             body,
		}
		if m.MessageID, err = cs.AddMessage(c.Request().Context(), m); err != nil {
			return err
		}
		tsh.Broadcast(m.Topic, appendChatMessageStream(m))

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			return r.TurboStream(c.Response(), replaceChatSendStream(topic, a))
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/instruments/"+name)
	}
}
