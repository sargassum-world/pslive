package pslive

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive/workers"
)

// Workers

type Worker func(ctx context.Context, s *Server) error

func periodicallyCleanupSessions(ctx context.Context, s *Server) error {
	const interval = 10 * time.Minute
	if err := s.Globals.Base.SessionsBacking.PeriodicallyCleanup(
		ctx, interval,
	); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "couldn't periodically clean up session store"))
	}
	return nil
}

func serveTSBroker(ctx context.Context, s *Server) error {
	if err := s.Globals.Base.TSBroker.Serve(ctx); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "turbo streams broker encountered error while serving"))
	}
	return nil
}

func serveVSBroker(ctx context.Context, s *Server) error {
	if err := s.Globals.VSBroker.Serve(ctx); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "video streams broker encountered error while serving"))
	}
	return nil
}

func establishPlanktoscopeConnections(ctx context.Context, s *Server) error {
	if err := workers.EstablishPlanktoscopeControllerConnections(
		ctx, s.Globals.Instruments, s.Globals.Planktoscopes,
	); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "couldn't establish planktoscope controller connections"))
	}
	return nil
}

func orchestrateInstrumentJobs(ctx context.Context, s *Server) error {
	if err := s.Globals.InstrumentJobs.Orchestrate(ctx); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "instrument job orchestrator encountered error while orchestrating"))
	}
	return nil
}

func startInstrumentJobs(ctx context.Context, s *Server) error {
	if err := workers.StartInstrumentJobs(
		ctx, s.Globals.Instruments, s.Globals.InstrumentJobs,
	); err != nil && err != context.Canceled {
		l := s.Globals.Base.Logger
		l.Error(errors.Wrap(err, "couldn't start automation jobs"))
	}
	return nil
}

func DefaultWorkers() []Worker {
	return []Worker{
		periodicallyCleanupSessions,
		serveTSBroker,
		serveVSBroker,
		establishPlanktoscopeConnections,
		orchestrateInstrumentJobs,
		startInstrumentJobs,
	}
}
