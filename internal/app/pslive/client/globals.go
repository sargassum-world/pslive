// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
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
)

type Globals struct {
	Config conf.Config
	Cache  clientcache.Cache
	DB     *database.DB

	Sessions        *session.Store
	SessionsBacking *sqlitestore.SqliteStore
	CSRFChecker     *session.CSRFTokenChecker
	Ory             *ory.Client
	AuthzChecker    *auth.AuthzChecker

	ACCancellers *actioncable.Cancellers
	TSSigner     turbostreams.Signer
	TSBroker     *turbostreams.Broker

	Instruments   *instruments.Store
	Planktoscopes *planktoscope.Orchestrator
	Presence      *presence.Store
	Chat          *chat.Store

	Logger godest.Logger
}

func NewGlobals(
	persistenceEmbeds database.Embeds, regoRoutesPackage string, regoModules []opa.Module,
	l godest.Logger,
) (g *Globals, err error) {
	g = &Globals{}
	g.Config, err = conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}
	if g.Cache, err = clientcache.NewRistrettoCache(g.Config.Cache); err != nil {
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
	tssConfig, err := turbostreams.GetSignerConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up turbo streams signer config")
	}
	g.TSSigner = turbostreams.NewSigner(tssConfig)
	g.TSBroker = turbostreams.NewBroker(l)

	g.Instruments = instruments.NewStore(g.DB)
	g.Planktoscopes = planktoscope.NewOrchestrator(l)
	g.Presence = presence.NewStore()
	g.Chat = chat.NewStore(g.DB)

	g.Logger = l
	return g, nil
}
