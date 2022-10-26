package handling

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type ChatMessageViewData struct {
	ID               int64
	Topic            string
	SendTime         time.Time
	SenderID         string
	SenderIdentifier string
	Body             string
}

func NewChatMessageViewData(m chat.Message) ChatMessageViewData {
	return ChatMessageViewData{
		ID:       m.ID,
		Topic:    m.Topic,
		SendTime: m.SendTime,
		SenderID: m.SenderID,
		Body:     m.Body,
	}
}

func AdaptChatMessages(
	ctx context.Context, messages []chat.Message, oc *ory.Client,
) (viewData []ChatMessageViewData, err error) {
	viewData = make([]ChatMessageViewData, len(messages))
	for i, message := range messages {
		viewData[i] = NewChatMessageViewData(message)
		if viewData[i].SenderIdentifier, err = oc.GetIdentifier(ctx, message.SenderID); err != nil {
			return nil, errors.Wrapf(
				err, "couldn't look up identifier of message sender %s", message.SenderID,
			)
		}
	}
	return viewData, nil
}

const (
	messagePartial = "shared/chat/message.partial.tmpl"
	sendPartial    = "shared/chat/send.partial.tmpl"
)

func appendChatMessageStream(m ChatMessageViewData) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionAppend,
		Target:   m.Topic,
		Template: messagePartial,
		Data: map[string]interface{}{
			"Message":          m,
			"AutoscrollOnLoad": true,
		},
	}
}

func replaceChatSendStream(topic string, authorizeSend bool, a auth.Auth) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   topic + "/send",
		Template: sendPartial,
		Data: map[string]interface{}{
			"Topic":         topic,
			"AuthorizeSend": authorizeSend,
			"FocusOnLoad":   true,
			"Auth":          a,
		},
	}
}

func HandleChatMessagesPost(
	r godest.TemplateRenderer, oc *ory.Client, azc *auth.AuthzChecker,
	tsh *turbostreams.Hub, cs *chat.Store,
) auth.HTTPHandlerFunc {
	sendT := sendPartial
	r.MustHave(sendT)
	return func(c echo.Context, a auth.Auth) (err error) {
		// Parse params
		name := c.Param("name")
		body := c.FormValue("body")
		topic := strings.TrimSuffix(c.Request().URL.Path, "/messages")

		// Run queries
		ctx := c.Request().Context()
		user, err := oc.GetIdentifier(ctx, a.Identity.User)
		if err != nil {
			return err
		}
		m := chat.Message{
			Topic:    topic + "/messages",
			SendTime: time.Now(),
			SenderID: a.Identity.User,
			Body:     body,
		}
		if m.ID, err = cs.AddMessage(ctx, m); err != nil {
			return err
		}
		mvd := NewChatMessageViewData(m)
		mvd.SenderIdentifier = user
		tsh.Broadcast(m.Topic, []turbostreams.Message{appendChatMessageStream(mvd)})

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			authorizeSend, err := azc.Allow(ctx, a, m.Topic, http.MethodPost, nil)
			if err != nil {
				return errors.Wrapf(err, "couldn't check authz for sending to chat for topic %s", topic)
			}
			return r.TurboStream(c.Response(), replaceChatSendStream(topic, authorizeSend, a))
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/instruments/"+name)
	}
}
