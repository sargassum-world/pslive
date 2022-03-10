// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/planktoscopes"
)

type Clients struct {
	Authn    *authn.Client
	Sessions *session.Client

	Planktoscopes *planktoscopes.Client
}

type Globals struct {
	Config  conf.Config
	Clients *Clients
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	g.Config, err = conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}
	g.Clients = &Clients{}

	authnConfig, err := authn.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn config")
	}
	g.Clients.Authn = authn.NewClient(authnConfig)

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Clients.Sessions = session.NewMemStoreClient(sessionsConfig)

	pcConfig, err := planktoscopes.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up planktoscopes config")
	}
	g.Clients.Planktoscopes = planktoscopes.NewClient(pcConfig, l)

	return g, nil
}
