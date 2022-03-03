// Package middleware provides generic middleware
package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// isContentType validates the Content-Type header matches the supplied
// contentType. That is, its type and subtype match.
func isContentType(h http.Header, contentType string) bool {
	// This is a copy of the isContentType function in github.com/gorilla/handlers
	ct := h.Get(echo.HeaderContentType)
	if i := strings.IndexRune(ct, ';'); i != -1 {
		ct = ct[0:i]
	}
	return ct == contentType
}

// RequireContentTypes returns an echo.MiddlewareFunc which validates the request content type is
// compatible with the contentTypes list. It writes a HTTP 415 error if that fails.
//
// Only PUT, POST, and PATCH requests are considered.
func RequireContentTypes(contentTypes ...string) echo.MiddlewareFunc {
	// This is an adaptation for echo of the RequireContentTypes function in
	// github.com/gorilla/handlers
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			if !(r.Method == http.MethodPut || r.Method == http.MethodPost || r.Method == http.MethodPatch) {
				return next(c)
			}

			for _, ct := range contentTypes {
				if isContentType(r.Header, ct) {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, fmt.Sprintf(
				"Unsupported content type %q; expected one of %q",
				r.Header.Get(echo.HeaderContentType), contentTypes,
			))
		}
	}
}
