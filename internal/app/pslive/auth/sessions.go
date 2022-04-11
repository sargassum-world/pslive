package auth

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

func GetWithoutRequest(s sessions.Session, ss session.Store) (a Auth, err error) {
	a.Identity, err = GetIdentity(s)
	if err != nil {
		return Auth{}, err
	}

	a.CSRF.Config = ss.CSRFOptions()
	a.CSRF.Behavior, err = GetCSRFBehavior(s)
	if err != nil {
		return Auth{}, err
	}
	// We don't provide an inline token here because it must be associated with a cookie, which must
	// come from an HTTP request.
	return a, nil
}

// HTTP

func Get(r *http.Request, s sessions.Session, ss session.Store) (a Auth, err error) {
	return GetFromRequest(r, s, ss)
}

func GetFromRequest(r *http.Request, s sessions.Session, ss session.Store) (a Auth, err error) {
	a, err = GetWithoutRequest(s, ss)
	if err != nil {
		return Auth{}, err
	}

	if a.CSRF.Behavior.InlineToken {
		a.CSRF.Token = csrf.Token(r)
	}
	return a, nil
}

func GetWithSession(
	r *http.Request, ss session.Store, l godest.Logger,
) (a Auth, s *sessions.Session, err error) {
	s, err = ss.Get(r)
	if err != nil {
		// If the user doesn't have a valid session, create one
		if s, err = ss.New(r); err != nil {
			// When an error is returned, a new (valid) session is still created
			l.Warnf("created new session to replace invalid session")
		}
		// We let the caller save the new session
	}
	a, err = Get(r, *s, ss)
	if err != nil {
		return Auth{}, s, err
	}
	return a, s, nil
}
