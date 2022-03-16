// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/session"

	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type Clients struct {
	Authn    *authn.Client
	Sessions *session.Client

	Instruments  *instruments.Client
	Planktoscope *planktoscope.Client
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

	instrumentsConfig, err := instruments.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up instruments config")
	}
	g.Clients.Instruments = instruments.NewClient(instrumentsConfig, l)

	planktoscopeConfig, err := planktoscope.GetConfig(instrumentsConfig.Instrument.Controller)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up planktoscope config")
	}
	g.Clients.Planktoscope, err = planktoscope.NewClient(planktoscopeConfig, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up planktoscope client")
	}

	return g, nil
}
