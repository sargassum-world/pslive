package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/app/pslive"
	"github.com/sargassum-world/pslive/internal/app/pslive/conf"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
)

const shutdownTimeout = 5 // sec

func overrideConfigDefaults() error {
	// TODO: override the defaults in a cleaner way
	if err := os.Setenv("SESSIONS_COOKIE_NOHTTPSONLY", "true"); err != nil {
		return err
	}
	if os.Getenv("DATABASE_URI") == "" {
		if err := os.Setenv("DATABASE_URI", "file:db-pslocal.sqlite3"); err != nil {
			return err
		}
	}
	if os.Getenv("AUTHN_NOAUTH") == "" {
		if err := os.Setenv("AUTHN_NOAUTH", "true"); err != nil {
			return err
		}
	}
	if os.Getenv("ORY_NOAUTH") == "" {
		if err := os.Setenv("ORY_NOAUTH", "true"); err != nil {
			return err
		}
	}
	if os.Getenv("PLANKTOSCOPE_MQTT_CLIENT") == "" {
		if err := os.Setenv(
			"PLANKTOSCOPE_MQTT_CLIENT", fmt.Sprintf("pslocal-%d", os.Getpid()),
		); err != nil {
			return err
		}
	}
	return nil
}

func initializeFromEmpty(ctx context.Context, server *pslive.Server) error {
	is := server.Globals.Instruments
	if instruments, err := server.Globals.Instruments.GetInstruments(
		ctx,
	); err != nil || len(instruments) > 0 {
		return errors.Wrap(err, "couldn't query for instruments")
	}

	iid, err := is.AddInstrument(ctx, instruments.Instrument{
		Name:    "planktoscope",
		AdminID: "admin",
	})
	if err != nil {
		return errors.Wrap(err, "couldn't add default instrument for local planktoscope")
	}
	if _, err = is.AddCamera(ctx, instruments.Camera{
		InstrumentID: iid,
		Enabled:      true,
		Name:         "preview",
		Description:  "The picamera preview stream",
		Protocol:     "mjpeg",
		URL:          "http://localhost:8000",
	}); err != nil {
		return errors.Wrap(err, "couldn't add default camera for local planktoscope")
	}
	controller := instruments.Controller{
		InstrumentID: iid,
		Enabled:      true,
		Name:         "controller",
		Description:  "The MQTT control API",
		Protocol:     "planktoscope-v2.3",
		URL:          "mqtt://localhost:1883",
	}
	controllerID, err := is.AddController(ctx, controller)
	if err != nil {
		return errors.Wrap(err, "couldn't add default controller for local planktoscope")
	}
	if err := server.Globals.Planktoscopes.Add(
		planktoscope.ClientID(controllerID), controller.URL,
	); err != nil {
		return errors.Wrap(err, "couldn't start mqtt client for local planktoscope")
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
	server, err := pslive.NewServer(
		config, append(pslive.DefaultWorkers(), initializeFromEmpty), e.Logger,
	)
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
