// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/authn"
	"github.com/sargassum-world/godest/clientcache"
	"github.com/sargassum-world/godest/database"
	"github.com/sargassum-world/godest/opa"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/session/sqlitestore"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

type BaseGlobals struct {
	Cache clientcache.Cache
	DB    *database.DB

	Sessions        *session.Store
	SessionsBacking *sqlitestore.SqliteStore
	CSRFChecker     *session.CSRFTokenChecker
	Authn           *authn.Client
	Ory             *ory.Client
	AuthzChecker    *auth.AuthzChecker

	ACCancellers *actioncable.Cancellers
	ACSigner     actioncable.Signer
	TSBroker     *turbostreams.Broker

	Logger godest.Logger
}

type Globals struct {
	Config conf.Config
	Base   *BaseGlobals

	Instruments    *instruments.Store
	Planktoscopes  *planktoscope.Orchestrator
	InstrumentJobs *instruments.JobOrchestrator

	Presence *presence.Store
	Chat     *chat.Store
	VSBroker *videostreams.Broker
}

func NewBaseGlobals(
	config conf.Config, persistenceEmbeds database.Embeds,
	regoRoutesPackage string, regoModules []opa.Module,
	l godest.Logger,
) (g *BaseGlobals, err error) {
	g = &BaseGlobals{}
	if g.Cache, err = clientcache.NewRistrettoCache(config.Cache); err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}
	storeConfig, err := database.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up persistent store config")
	}
	g.DB = database.NewDB(
		storeConfig,
		database.WithPrepareConnQueries(persistenceEmbeds.PrepareConnQueriesFS),
	)

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Sessions, g.SessionsBacking = sqlitestore.NewStore(g.DB, sessionsConfig)
	g.CSRFChecker = session.NewCSRFTokenChecker(sessionsConfig)

	authnConfig, err := authn.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up local authn config")
	}
	g.Authn = authn.NewClient(authnConfig)
	oryConfig, err := ory.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up ory config")
	}
	g.Ory = ory.NewClient(oryConfig, g.Cache, l)
	opc, err := opa.NewClient(regoRoutesPackage, opa.Modules(regoModules...))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up opa client")
	}
	g.AuthzChecker = auth.NewAuthzChecker(g.DB, opc)

	g.ACCancellers = actioncable.NewCancellers()
	acsConfig, err := actioncable.GetSignerConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up action cable signer config")
	}
	g.ACSigner = actioncable.NewSigner(acsConfig)
	g.TSBroker = turbostreams.NewBroker(l)

	g.Logger = l
	return g, nil
}

func NewGlobals(
	config conf.Config, persistenceEmbeds database.Embeds,
	regoRoutesPackage string, regoModules []opa.Module,
	l godest.Logger,
) (g *Globals, err error) {
	g = &Globals{
		Config: config,
	}
	g.Base, err = NewBaseGlobals(config, persistenceEmbeds, regoRoutesPackage, regoModules, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up base globals")
	}

	g.Instruments = instruments.NewStore(g.Base.DB)
	g.Planktoscopes = planktoscope.NewOrchestrator(l)
	instrumentControllerActionRunners := instruments.NewControllerActionRunnerStore(
		g.Instruments,
		map[string]instruments.ControllerActionRunnerGetter{
			"planktoscope-v2.3": NewPlanktoScopeControllerActionRunnerGetter(g.Planktoscopes),
		},
	)
	g.InstrumentJobs = instruments.NewJobOrchestrator(map[string]instruments.ActionHandler{
		"sleep":      instruments.HandleSleepAction,
		"controller": instrumentControllerActionRunners.HandleControllerAction,
	}, l)

	g.Presence = presence.NewStore()
	g.Chat = chat.NewStore(g.Base.DB)
	g.VSBroker = videostreams.NewBroker(l)

	return g, nil
}
