// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/godest"
)

const (
	AppURLPrefix    = "/app/"
	StaticURLPrefix = "/static/"
	FontsURLPrefix  = "/fonts/"
)

type TemplatedHandlers struct {
	r godest.TemplateRenderer
}

func NewTemplated(r godest.TemplateRenderer) *TemplatedHandlers {
	return &TemplatedHandlers{
		r: r,
	}
}

func (h *TemplatedHandlers) Register(er godest.EchoRouter) {
	er.GET(AppURLPrefix+"app.webmanifest", h.getWebmanifest())
	er.GET(AppURLPrefix+"offline", h.getOffline())
}

func RegisterStatic(er godest.EchoRouter, em godest.Embeds) {
	const (
		day  = 24 * time.Hour
		week = 7 * day
		year = 365 * day
	)

	// TODO: serve sw.js with an ETag!
	er.GET("/sw.js", echo.WrapHandler(godest.HandleFS("/", em.AppFS, week)))
	er.GET("/favicon.ico", echo.WrapHandler(godest.HandleFS("/", em.StaticFS, week)))
	er.GET(FontsURLPrefix+"*", echo.WrapHandler(godest.HandleFS(FontsURLPrefix, em.FontsFS, year)))
	er.GET(
		StaticURLPrefix+"*",
		echo.WrapHandler(godest.HandleFSFileRevved(StaticURLPrefix, em.StaticHFS)),
	)
	er.GET(AppURLPrefix+"*", echo.WrapHandler(godest.HandleFSFileRevved(AppURLPrefix, em.AppHFS)))
}
