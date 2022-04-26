// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/clientcache"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

type Globals struct {
	Config conf.Config
	Cache  clientcache.Cache

	Sessions    session.Store
	CSRFChecker *session.CSRFTokenChecker
	Ory         *ory.Client

	ACCancellers *actioncable.Cancellers
	TSSigner     turbostreams.Signer
	TSBroker     *turbostreams.Broker

	Instruments   *instruments.Client
	Planktoscopes map[string]*planktoscope.Client
	Presence      *presence.Store

	Logger godest.Logger
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	g.Config, err = conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}
	if g.Cache, err = clientcache.NewRistrettoCache(g.Config.Cache); err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Sessions = session.NewMemStore(sessionsConfig)
	g.CSRFChecker = session.NewCSRFTokenChecker(sessionsConfig)
	oryConfig, err := ory.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up ory config")
	}
	g.Ory = ory.NewClient(oryConfig, g.Cache, l)

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
	g.Presence = presence.NewStore()

	g.Logger = l
	return g, nil
}
