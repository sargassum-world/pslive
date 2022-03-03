// Package pslive provides the Planktoscope Live server.
package pslive

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/httperr"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type ErrorData struct {
	Code     int
	Error    httperr.DescriptiveError
	Messages []string
}

func NewHTTPErrorHandler(tr godest.TemplateRenderer, sc *sessions.Client) echo.HTTPErrorHandler {
	tr.MustHave("app/httperr.page.tmpl")
	return func(err error, c echo.Context) {
		c.Logger().Error(err)

		// Check authentication & authorization
		a, sess, serr := auth.GetWithSession(c, sc)
		if serr != nil {
			c.Logger().Error(errors.Wrap(serr, "couldn't get session+auth in error handler"))
		}

		// Process error code
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		errorData := ErrorData{
			Code:  code,
			Error: httperr.Describe(code),
		}

		// Consume & save session
		if sess != nil {
			messages, merr := session.GetErrorMessages(sess)
			if merr != nil {
				c.Logger().Error(errors.Wrap(
					merr, "couldn't get error messages from session in error handler",
				))
			}
			errorData.Messages = messages
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				c.Logger().Error(errors.Wrap(serr, "couldn't save session in error handler"))
			}
		}

		// Produce output
		tr.SetUncacheable(c.Response().Header())
		if perr := tr.Page(
			c.Response(), c.Request(), code, "app/httperr.page.tmpl", errorData, a,
		); perr != nil {
			c.Logger().Error(errors.Wrap(perr, "couldn't render error page in error handler"))
		}
	}
}

func NewCSRFErrorHandler(
	tr godest.TemplateRenderer, l echo.Logger, sc *sessions.Client,
) http.HandlerFunc {
	tr.MustHave("app/httperr.page.tmpl")
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error(csrf.FailureReason(r))
		// Check authentication & authorization
		sess, serr := session.Get(r, sc.Config.CookieName, sc.Store)
		if serr != nil {
			l.Error(errors.Wrap(serr, "couldn't get session in error handler"))
		}
		var a auth.Auth
		if sess != nil {
			a, serr = auth.GetFromRequest(r, *sess, sc)
			if serr != nil {
				l.Error(errors.Wrap(serr, "couldn't get auth in error handler"))
			}
		}

		// Generate error code
		code := http.StatusForbidden
		errorData := ErrorData{
			Code:  code,
			Error: httperr.Describe(code),
			Messages: []string{
				fmt.Sprintf(
					"%s. If you disabled Javascript after signing in, "+
						"please clear your cookies for this site and sign in again.",
					csrf.FailureReason(r).Error(),
				),
			},
		}

		// Produce output
		tr.SetUncacheable(w.Header())
		if rerr := tr.Page(w, r, code, "app/httperr.page.tmpl", errorData, a); rerr != nil {
			l.Error(errors.Wrap(rerr, "couldn't render error page in error handler"))
		}
	}
}
