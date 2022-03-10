package auth

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
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
		return Identity{}, nil
	}

	rawIdentity, ok := s.Values["identity"]
	if !ok {
		// A zero value for Identity indicates that the session has no identity associated with it
		return Identity{}, nil
	}
	identity, ok = rawIdentity.(Identity)
	if !ok {
		return Identity{}, errors.Errorf("unexpected type for field identity in session")
	}
	return identity, nil
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
		return CSRFBehavior{}, nil
	}

	rawBehavior, ok := s.Values["csrfBehavior"]
	if !ok {
		// By default, HTML responses won't inline the CSRF input fields (so responses can be cached),
		// because the app only allows POST requests after user authentication. This default behavior
		// can be overridden, e.g. on the login form for user authentication, with OverrideCSRFInlining.
		return CSRFBehavior{}, nil
	}
	behavior, ok = rawBehavior.(CSRFBehavior)
	if !ok {
		return CSRFBehavior{}, errors.Errorf("unexpected type for field csrfBehavior in session")
	}
	return behavior, nil
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

func Get(r *http.Request, s sessions.Session, sc *session.Client) (a Auth, err error) {
	return GetFromRequest(r, s, sc)
}

func GetFromRequest(r *http.Request, s sessions.Session, sc *session.Client) (a Auth, err error) {
	a.Identity, err = GetIdentity(s)
	if err != nil {
		return Auth{}, err
	}

	a.CSRF.Config = sc.Config.CSRFOptions
	a.CSRF.Behavior, err = GetCSRFBehavior(s)
	if err != nil {
		return Auth{}, err
	}
	if a.CSRF.Behavior.InlineToken {
		a.CSRF.Token = csrf.Token(r)
	}
	return a, nil
}

func GetWithSession(
	r *http.Request, sc *session.Client, l godest.Logger,
) (a Auth, s *sessions.Session, err error) {
	s, err = sc.Get(r)
	if err != nil {
		// If the user doesn't have a valid session, create one
		if s, err = sc.New(r); err != nil {
			// When an error is returned, a new (valid) session is still created
			l.Warnf("created new session to replace invalid session")
		}
		// We let the caller save the new session
	}
	a, err = Get(r, *s, sc)
	if err != nil {
		return Auth{}, s, err
	}
	return a, s, nil
}
