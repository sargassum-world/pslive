// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

type Globals struct {
	Config conf.Config

	Sessions    session.Store
	CSRFChecker *session.CSRFTokenChecker
	Authn       *authn.Client

	ACCancellers *actioncable.Cancellers
	TSSigner     turbostreams.Signer
	TSBroker     *turbostreams.Broker

	Instruments   *instruments.Client
	Planktoscopes map[string]*planktoscope.Client

	Logger godest.Logger
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	g.Config, err = conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Sessions = session.NewMemStore(sessionsConfig)
	g.CSRFChecker = session.NewCSRFTokenChecker(sessionsConfig)
	authnConfig, err := authn.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn config")
	}
	g.Authn = authn.NewClient(authnConfig)

	g.ACCancellers = actioncable.NewCancellers()
	tssConfig, err := turbostreams.GetSignerConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up turbo streams signer config")
	}
	g.TSSigner = turbostreams.NewSigner(tssConfig)
	g.TSBroker = turbostreams.NewBroker(l)

	instrumentsConfig, err := instruments.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up instruments config")
	}
	g.Instruments = instruments.NewClient(instrumentsConfig, l)
	planktoscopeConfig, err := planktoscope.GetConfig(instrumentsConfig.Instrument.Controller)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up planktoscope config")
	}
	g.Planktoscopes = make(map[string]*planktoscope.Client)
	g.Planktoscopes[instrumentsConfig.Instrument.Controller], err = planktoscope.NewClient(
		planktoscopeConfig, l,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up planktoscope client")
	}

	g.Logger = l
	return g, nil
}
