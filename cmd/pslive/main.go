package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/sargassum-world/pslive/internal/app/pslive"
)

const port = 3000

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} ${method} ${uri} (${bytes_in}b) => " +
			"(${bytes_out}b after ${latency_human}) ${status} ${error}\n",
	}))
	e.Logger.SetLevel(log.DEBUG)

	// Prepare server
	s, err := pslive.NewServer(e)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
	s.Register(e)

	// Start server
	go s.RunBackgroundWorkers()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
