package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/pslive/internal/app/pslive"
)

const shutdownTimeout = 5 // sec

func main() {
	// Prepare server
	e := echo.New()
	server, err := pslive.NewServer(e.Logger)
	if err != nil {
		e.Logger.Fatal(err)
	}
	server.Register(e)

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
		closeErr := server.Close(e)
		if closeErr != nil {
			e.Logger.Error(closeErr)
		}
	}
	e.Logger.Info("finished shutdown")
}
