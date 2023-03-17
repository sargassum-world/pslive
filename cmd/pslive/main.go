package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive"
	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
)

const shutdownTimeout = 5 // sec

func overrideConfigDefaults() error {
	// TODO: override the defaults in a cleaner way
	if os.Getenv("ORY_NOAUTH") == "" || os.Getenv("ORY_NOAUTH") == "false" {
		// If we have Ory for authentication, we want to discourage use of the local admin account
		if err := os.Setenv("AUTHN_NOAUTH", "true"); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	e := echo.New()

	if err := overrideConfigDefaults(); err != nil {
		e.Logger.Fatal(err, "couldn't override application config defaults")
	}

	// Get config
	config, err := conf.GetConfig()
	if err != nil {
		e.Logger.Fatal(err, "couldn't set up application config")
	}

	// Prepare server
	server, err := pslive.NewServer(config, pslive.DefaultWorkers(), e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}
	if err = server.Register(e); err != nil {
		e.Logger.Fatal(err)
	}

	// Run server
	ctxRun, cancelRun := signal.NotifyContext(
		context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT,
	)
	go func() {
		if err = server.Run(e); err != nil {
			e.Logger.Error(err)
		}
		cancelRun()
	}()
	<-ctxRun.Done()
	cancelRun()

	// Shut down server
	ctxShutdown, cancelShutdown := context.WithTimeout(
		context.Background(), shutdownTimeout*time.Second,
	)
	defer cancelShutdown()
	e.Logger.Infof("attempting to shut down gracefully within %d sec", shutdownTimeout)
	if err := server.Shutdown(ctxShutdown, e); err != nil {
		e.Logger.Warn("forcibly closing http server due to failure of graceful shutdown")
		if closeErr := server.Close(e); closeErr != nil {
			e.Logger.Error(closeErr)
		}
	}
	e.Logger.Info("finished shutdown")
}
