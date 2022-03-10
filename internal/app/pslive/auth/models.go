// Package auth provides application-level standardization for authentication
package auth

import (
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Identity struct {
	Authenticated bool
	User          string
}

type CSRFBehavior struct {
	InlineToken bool
}

type CSRF struct {
	Config   session.CSRFOptions
	Behavior CSRFBehavior
	Token    string
}

type Auth struct {
	Identity Identity
	CSRF     CSRF
}
