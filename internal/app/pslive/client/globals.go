// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/authn"
	"github.com/sargassum-world/pslive/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/pkg/godest"
)

type Clients struct {
	Authn         *authn.Client
	Sessions      *sessions.Client
}

type Globals struct {
	Config  conf.Config
	Clients *Clients
}

func NewGlobals(l godest.Logger) (*Globals, error) {
	config, err := conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}

	authnClient, err := authn.NewClient(l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn client")
	}

	sessionsClient, err := sessions.NewMemStoreClient(l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions client")
	}

	return &Globals{
		Config: config,
		Clients: &Clients{
			Authn:         authnClient,
			Sessions:      sessionsClient,
		},
	}, nil
}
