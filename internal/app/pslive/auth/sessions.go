package auth

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	sessionsc "github.com/sargassum-world/pslive/internal/clients/sessions"
)

// Identity

func SetIdentity(s *sessions.Session, username string) {
	identity := Identity{
		Authenticated: username != "",
		User:          username,
	}
	s.Values["identity"] = identity
	gob.Register(identity)
}

func GetIdentity(s sessions.Session) (identity Identity, err error) {
	if s.IsNew {
		return
	}

	rawIdentity, ok := s.Values["identity"]
	if !ok {
		// A zero value for Identity indicates that the session has no identity associated with it
		return
	}
	identity, ok = rawIdentity.(Identity)
	if !ok {
		err = fmt.Errorf("unexpected type for field identity in session")
		return
	}
	return
}

// CSRF

func SetCSRFBehavior(s *sessions.Session, inlineToken bool) {
	behavior := CSRFBehavior{
		InlineToken: inlineToken,
	}
	s.Values["csrfBehavior"] = behavior
	gob.Register(behavior)
}

func GetCSRFBehavior(s sessions.Session) (behavior CSRFBehavior, err error) {
	if s.IsNew {
		return
	}

	rawBehavior, ok := s.Values["csrfBehavior"]
	if !ok {
		// By default, HTML responses won't inline the CSRF input fields (so responses can be cached),
		// because the app only allows POST requests after user authentication. This default behavior
		// can be overridden, e.g. on the login form for user authentication, with OverrideCSRFInlining.
		return
	}
	behavior, ok = rawBehavior.(CSRFBehavior)
	if !ok {
		err = fmt.Errorf("unexpected type for field csrfBehavior in session")
		return
	}
	return
}

func (c *CSRF) SetInlining(r *http.Request, inlineToken bool) {
	c.Behavior.InlineToken = inlineToken
	if c.Behavior.InlineToken {
		c.Token = csrf.Token(r)
	} else {
		c.Token = ""
	}
}

// Access

func Get(c echo.Context, s sessions.Session, sc *sessionsc.Client) (a Auth, err error) {
	return GetFromRequest(c.Request(), s, sc)
}

func GetFromRequest(r *http.Request, s sessions.Session, sc *sessionsc.Client) (a Auth, err error) {
	a.Identity, err = GetIdentity(s)
	if err != nil {
		return
	}

	a.CSRF.Config = sc.Config.CSRFOptions
	a.CSRF.Behavior, err = GetCSRFBehavior(s)
	if err != nil {
		return
	}
	if a.CSRF.Behavior.InlineToken {
		a.CSRF.Token = csrf.Token(r)
	}
	return
}

func GetWithSession(c echo.Context, sc *sessionsc.Client) (a Auth, s *sessions.Session, err error) {
	s, err = sc.Get(c)
	if err != nil {
		return Auth{}, nil, err
	}
	a, err = Get(c, *s, sc)
	if err != nil {
		return Auth{}, s, err
	}

	return
}
