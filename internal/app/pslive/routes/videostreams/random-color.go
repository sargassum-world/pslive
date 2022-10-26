package videostreams

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/pslive/internal/clients/mjpeg"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Color generation

func newRandomColor() color.RGBA {
	const ceiling = 256
	//nolint:gosec // We don't need cryptographically-secure random number generation here
	return color.RGBA{
		R: uint8(rand.Intn(ceiling)),
		G: uint8(rand.Intn(ceiling)),
		B: uint8(rand.Intn(ceiling)),
		A: ceiling - 1,
	}
}

// Image generation

func newUniformImage(width, height int, co color.Color) image.Image {
	im := image.NewRGBA(image.Rect(0, 0, width, height))
	fillImage(im, co)
	return im
}

func fillImage(im draw.Image, co color.Color) {
	draw.Draw(im, im.Bounds(), &image.Uniform{co}, image.Point{}, draw.Src)
}

func newCurrentFrame(im image.Image) videostreams.ImageFrame {
	return videostreams.ImageFrame{
		Timestamp: time.Now(),
		Data:      im,
	}
}

// Query Parameter Parsing

func parseIntParam(raw, name string, defaultValue int) (parsed int, err error) {
	parsed = defaultValue
	if raw != "" {
		if parsed, err = strconv.Atoi(raw); err != nil {
			return 0, errors.Wrapf(err, "invalid %s parameter %s", name, raw)
		}
	}
	return parsed, err
}

func parseImageDimensions(rawWidth, rawHeight string) (width, height int, err error) {
	width = 800
	if width, err = parseIntParam(rawWidth, "width", width); err != nil {
		return 0, 0, err
	}

	height = 600
	if height, err = parseIntParam(rawHeight, "height", height); err != nil {
		return 0, 0, err
	}

	return width, height, nil
}

// Handlers

func (h *Handlers) HandleRandomColorFrameGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		width, height, err := parseImageDimensions(c.QueryParam("width"), c.QueryParam("height"))
		if err != nil {
			return errors.Wrap(err, "couldn't parse image dimension query parameter")
		}
		const defaultQuality = 20
		quality, err := parseIntParam(c.QueryParam("quality"), "quality", defaultQuality)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image quality query parameter")
		}

		// Generate data
		frame := newCurrentFrame(newUniformImage(width, height, newRandomColor()))
		frame.JPEGQuality = quality
		frameJPEG, err := frame.JPEG()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}

		// Produce output
		const base = 10
		c.Response().Header().Set("X-Timestamp", strconv.FormatInt(frame.Time().UnixMilli(), base))
		return c.Blob(http.StatusOK, "image/jpeg", frameJPEG)
	}
}

func (h *Handlers) HandleRandomColorStreamGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		width, height, err := parseImageDimensions(c.QueryParam("width"), c.QueryParam("height"))
		if err != nil {
			return errors.Wrap(err, "couldn't parse image dimension query parameter")
		}
		const defaultQuality = 20
		quality, err := parseIntParam(c.QueryParam("quality"), "quality", defaultQuality)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image quality query parameter")
		}

		frames := make(chan mjpeg.EncodableFrame)
		// TODO: also make it dependent on the server's context!
		eg, egctx := errgroup.WithContext(c.Request().Context())
		eg.Go(func() error {
			const interval = 1 * time.Second
			return handling.Repeat(egctx, interval, func() (done bool, err error) {
				// Generate data
				frame := newCurrentFrame(newUniformImage(width, height, newRandomColor()))
				frame.JPEGQuality = quality
				frames <- frame
				return false, nil
			})
		})
		eg.Go(func() error {
			// Produce output
			return mjpeg.SendStream(egctx, c.Response().Writer, frames)
		})
		if err := handling.Except(eg.Wait(), context.Canceled, syscall.EPIPE); err != nil {
			return err
		}
		close(frames)
		return nil
	}
}
