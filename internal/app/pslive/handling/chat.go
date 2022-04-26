package handling

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

const (
	messagePartial = "shared/chat/message.partial.tmpl"
	sendPartial    = "shared/chat/send.partial.tmpl"
)

func appendChatMessageStream(topic, userID, userIdentifier, message string) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionAppend,
		Target:   topic + "/messages",
		Template: messagePartial,
		Data: map[string]interface{}{
			"UserID":         userID,
			"UserIdentifier": userIdentifier,
			"Message":        message,
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
	r godest.TemplateRenderer, oc *ory.Client, tsh *turbostreams.MessagesHub,
) auth.HTTPHandlerFunc {
	sendT := sendPartial
	r.MustHave(sendT)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")
		message := c.FormValue("message")
		topic := strings.TrimSuffix(c.Request().URL.Path, "/messages")

		// Run queries
		user, err := oc.GetIdentifier(c.Request().Context(), a.Identity.User)
		if err != nil {
			return err
		}
		tsh.Broadcast(topic+"/messages", appendChatMessageStream(topic, a.Identity.User, user, message))

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			message := replaceChatSendStream(topic, a)
			return r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/instruments/"+name)
	}
}
