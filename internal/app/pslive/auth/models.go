// Package auth provides application-level standardization for authentication
package auth

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/opa"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/pslive/internal/clients/ory"
)

type Auth struct {
	Identity       Identity
	Authorizations interface{}
	CSRF           CSRF
}

func RegisterGobTypes() {
	gob.Register(Identity{})
	gob.Register(CSRF{})
	gob.Register(CSRFBehavior{})
}

// Identity

type Identity struct {
	Authenticated bool
	User          ory.IdentityID
}

func (i Identity) NewSubject() opa.Subject {
	return opa.NewSubject(string(i.User), i.Authenticated)
}

func SetIdentity(s *sessions.Session, id ory.IdentityID) {
	s.Values["identity"] = Identity{
		Authenticated: id != "",
		User:          id,
	}
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

type CSRFBehavior struct {
	InlineToken bool
}

type CSRF struct {
	Config   session.CSRFOptions
	Behavior CSRFBehavior
	Token    string
}

func SetCSRFBehavior(s *sessions.Session, inlineToken bool) {
	behavior := CSRFBehavior{
		InlineToken: inlineToken,
	}
	s.Values["csrfBehavior"] = behavior
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
