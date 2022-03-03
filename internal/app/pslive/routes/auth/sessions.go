package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/sessions"
)

type CSRFData struct {
	HeaderName string `json:"headerName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Token      string `json:"token,omitempty"`
}

func (h *Handlers) HandleCSRFGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		sess, err := h.sc.Get(c)
		if err != nil {
			return err
		}
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		// Produce output
		h.r.SetUncacheable(c.Response().Header())
		return c.JSON(http.StatusOK, CSRFData{
			HeaderName: h.sc.Config.CSRFOptions.HeaderName,
			FieldName:  h.sc.Config.CSRFOptions.FieldName,
			Token:      csrf.Token(c.Request()),
		})
	}
}

type LoginData struct {
	NoAuth        bool
	ReturnURL     string
	ErrorMessages []string
}

func (h *Handlers) HandleLoginGet() auth.AuthAwareHandler {
	t := "auth/login.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Check authentication & authorization
		sess, err := h.sc.Get(c)
		if err != nil {
			return err
		}

		// Consume & save session
		errorMessages, err := session.GetErrorMessages(sess)
		if err != nil {
			return err
		}
		loginData := LoginData{
			NoAuth:        h.ac.Config.NoAuth,
			ReturnURL:     c.QueryParam("return"),
			ErrorMessages: errorMessages,
		}
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		// Add non-persistent overrides of session data
		a.CSRF.SetInlining(c.Request(), true)

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, loginData, a)
	}
}

func sanitizeReturnURL(returnURL string) (*url.URL, error) {
	u, err := url.ParseRequestURI(returnURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func handleAuthenticationSuccess(
	c echo.Context, username, returnURL string, omitCSRFToken bool, sc *sessions.Client,
) error {
	// Update session
	sess, err := sc.Regenerate(c)
	if err != nil {
		return err
	}
	auth.SetIdentity(sess, username)
	// This allows client-side Javascript to specify for server-side session data that we only need
	// to provide CSRF tokens through the /csrf route and we can omit them from HTML response
	// bodies, in order to make HTML responses cacheable.
	auth.SetCSRFBehavior(sess, !omitCSRFToken)
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	// Redirect user
	u, err := sanitizeReturnURL(returnURL)
	if err != nil {
		// TODO: log the error, too
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Redirect(http.StatusSeeOther, u.String())
}

func handleAuthenticationFailure(c echo.Context, returnURL string, sc *sessions.Client) error {
	// Update session
	sess, serr := sc.Get(c)
	if serr != nil {
		return serr
	}
	session.AddErrorMessage(sess, "Could not log in!")
	auth.SetIdentity(sess, "")
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	// Redirect user
	u, err := sanitizeReturnURL(returnURL)
	if err != nil {
		// TODO: log the error, too
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	r := url.URL{Path: "/login"}
	q := r.Query()
	q.Set("return", u.String())
	r.RawQuery = q.Encode()
	return c.Redirect(http.StatusSeeOther, r.String())
}

func (h *Handlers) HandleSessionsPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		state := c.FormValue("state")

		// Run queries
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid session %s", state,
			))
		case "authenticated":
			username := c.FormValue("username")
			password := c.FormValue("password")
			returnURL := c.FormValue("return")
			omitCSRFToken := strings.ToLower(c.FormValue("omit-csrf-token")) == "true"

			// TODO: add session attacks detection. Refer to the "Session Attacks Detection" section of
			// the OWASP Session Management Cheat Sheet

			// Check authentication
			identified, err := h.ac.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if !identified {
				return handleAuthenticationFailure(c, returnURL, h.sc)
			}
			return handleAuthenticationSuccess(c, username, returnURL, omitCSRFToken, h.sc)
		case "unauthenticated":
			// TODO: add a client-side controller to automatically submit a logout request after the
			// idle timeout expires, and display an inactivity logout message
			sess, err := h.sc.Invalidate(c)
			if err != nil {
				return err
			}
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				return err
			}
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/")
	}
}
