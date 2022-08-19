package pslive

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/httperr"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
)

type ErrorData struct {
	Code     int
	Error    httperr.DescriptiveError
	Messages []string
}

func NewHTTPErrorHandler(tr godest.TemplateRenderer, ss session.Store) echo.HTTPErrorHandler {
	tr.MustHave("app/httperr.page.tmpl")
	return func(err error, c echo.Context) {
		c.Logger().Error(err)

		// Check authentication & authorization
		a, sess, serr := auth.GetFromRequest(c.Request(), ss, c.Logger())
		if serr != nil {
			c.Logger().Error(errors.Wrap(serr, "couldn't get auth in error handler"))
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

		// Produce output
		if perr := tr.Page(
			c.Response(), c.Request(), code, "app/httperr.page.tmpl", errorData, a,
			godest.WithUncacheable(),
		); perr != nil {
			c.Logger().Error(errors.Wrap(perr, "couldn't render error page in error handler"))
		}
	}
}

func NewCSRFErrorHandler(
	tr godest.TemplateRenderer, l echo.Logger, ss session.Store,
) http.HandlerFunc {
	tr.MustHave("app/httperr.page.tmpl")
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error(csrf.FailureReason(r))
		// Check authentication & authorization
		a, sess, serr := auth.GetFromRequest(r, ss, l)
		if serr != nil {
			l.Error(errors.Wrap(serr, "couldn't get auth in error handler"))
		}
		// Save the session in case it was freshly generated
		if err := sess.Save(r, w); err != nil {
			l.Error(errors.Wrap(serr, "couldn't save session in error handler"))
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
		if rerr := tr.Page(
			w, r, code, "app/httperr.page.tmpl", errorData, a,
			godest.WithUncacheable(),
		); rerr != nil {
			l.Error(errors.Wrap(rerr, "couldn't render error page in error handler"))
		}
	}
}
