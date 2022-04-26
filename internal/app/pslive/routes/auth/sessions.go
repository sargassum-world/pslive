package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type CSRFData struct {
	HeaderName string `json:"headerName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Token      string `json:"token,omitempty"`
}

func (h *Handlers) HandleCSRFGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Produce output
		godest.WithUncacheable()(c.Response().Header())
		return c.JSON(http.StatusOK, CSRFData{
			HeaderName: h.ss.CSRFOptions().HeaderName,
			FieldName:  h.ss.CSRFOptions().FieldName,
			Token:      csrf.Token(c.Request()),
		})
	}
}

type LoginData struct {
	NoAuth         bool
	ReturnURL      string
	ErrorMessages  []string
	OryFlow        string
	OryCSRF        string
	OryRegisterURL string
	OryRecoverURL  string
	UserIdentifier string
}

func (h *Handlers) HandleLoginGet() auth.HTTPHandlerFuncWithSession {
	t := "auth/login.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth, sess *sessions.Session) error {
		// Consume & save session
		errorMessages, err := session.GetErrorMessages(sess)
		if err != nil {
			return err
		}
		if serr := sess.Save(c.Request(), c.Response()); serr != nil {
			return serr
		}

		flow, cookie, err := h.oc.InitializeLoginFlow(c.Request().Context())
		if err != nil {
			return err
		}
		cookie.Domain = ""
		// TODO: adjust the Secure field based on session store config options
		c.SetCookie(cookie)

		// Make login page
		loginData := LoginData{
			ReturnURL:     c.QueryParam("return"),
			ErrorMessages: errorMessages,
			OryFlow:       flow.Id,
		}
		for _, node := range flow.Ui.Nodes {
			inputAttrs := node.Attributes.UiNodeInputAttributes
			if inputAttrs != nil && inputAttrs.Name == "csrf_token" {
				if csrfToken, ok := inputAttrs.Value.(string); ok {
					loginData.OryCSRF = csrfToken
				}
			}
		}
		if loginData.OryRegisterURL, err = h.oc.GetPath(
			c.Request().Context(), "V0alpha2ApiService.GetSelfServiceRegistrationFlow",
			"/ui/registration",
		); err != nil {
			return err
		}
		if loginData.OryRecoverURL, err = h.oc.GetPath(
			c.Request().Context(), "V0alpha2ApiService.GetSelfServiceRecoveryFlow", "/ui/recovery",
		); err != nil {
			return err
		}
		if a.Identity.Authenticated {
			if loginData.UserIdentifier, err = h.oc.GetIdentifier(
				c.Request().Context(), a.Identity.User,
			); err != nil {
				return err
			}
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
	c echo.Context, identifier, returnURL string, omitCSRFToken bool, ss session.Store,
) error {
	// Update session
	sess, err := ss.Get(c.Request())
	if err != nil {
		return err
	}
	session.Regenerate(sess)
	auth.SetIdentity(sess, identifier)
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

func handleAuthenticationFailure(c echo.Context, returnURL string, ss session.Store) error {
	// Update session
	sess, serr := ss.Get(c.Request())
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
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	r := url.URL{Path: "/login"}
	q := r.Query()
	q.Set("return", u.String())
	r.RawQuery = q.Encode()
	return c.Redirect(http.StatusSeeOther, r.String())
}

func handleLogin(c echo.Context, oc *ory.Client, ss session.Store, l godest.Logger) error {
	// Parse params
	identifier := c.FormValue("identifier")
	password := c.FormValue("password")
	returnURL := c.FormValue("return")
	omitCSRFToken := strings.ToLower(c.FormValue("omit-csrf-token")) == "true"
	oryFlow := c.FormValue("ory-flow")
	oryCSRFToken := c.FormValue("ory-csrf-token")

	// Run queries
	login, cookies, err := oc.SubmitLoginFlow(
		c.Request().Context(), oryFlow, oryCSRFToken, identifier, password, c.Request().Cookies(),
	)
	if err != nil {
		l.Error(errors.Wrapf(err, "login failed for identifier %s", identifier))
		return handleAuthenticationFailure(c, returnURL, ss)
	}
	if login == nil || login.Session.Active == nil || !(*login.Session.Active) {
		l.Warnf("login failed for identifier %s", identifier)
		return handleAuthenticationFailure(c, returnURL, ss)
	}
	for _, cookie := range cookies {
		cookie.Domain = ""
		// TODO: adjust the Secure field based on session store config options
		c.SetCookie(cookie)
	}

	// TODO: add session attacks detection. Refer to the "Session Attacks Detection" section of
	// the OWASP Session Management Cheat Sheet

	return handleAuthenticationSuccess(c, login.Session.Identity.Id, returnURL, omitCSRFToken, ss)
}

func handleLogout(
	c echo.Context, oc *ory.Client, ss session.Store, acc *actioncable.Cancellers, ps *presence.Store,
) error {
	// Invalidate the session cookie
	// TODO: add a client-side controller to automatically submit a logout request after the
	// idle timeout expires, and display an inactivity logout message
	sess, err := ss.Get(c.Request())
	if err != nil {
		return err
	}
	ps.Forget(sess.ID)
	acc.Cancel(sess.ID)
	session.Invalidate(sess)
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	// Perform Ory Kratos logout
	cookies, err := oc.PerformLogout(c.Request().Context(), c.Request().Cookies())
	if err != nil {
		return err
	}
	for _, cookie := range cookies {
		cookie.Domain = ""
		if strings.HasPrefix(cookie.Name, "ory_session_") {
			cookie.MaxAge = -1
		}
		// TODO: adjust the Secure field based on session store config options
		c.SetCookie(cookie)
	}

	// Redirect user
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handlers) HandleSessionsPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		state := c.FormValue("state")
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid session %s", state,
			))
		case "authenticated":
			return handleLogin(c, h.oc, h.ss, h.l)
		case "unauthenticated":
			return handleLogout(c, h.oc, h.ss, h.acc, h.ps)
		}
	}
}
