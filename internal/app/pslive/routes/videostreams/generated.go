package videostreams

import (
	"context"
	"image"
	"image/color"
	"math/rand"
	"net/http"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"

	"github.com/sargassum-world/pslive/internal/clients/mjpeg"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Color generation

type animatedValue struct {
	value      uint8
	increasing bool
	reversed   bool
	floor      uint8
	ceiling    uint8
}

func (v animatedValue) step(size uint8) animatedValue {
	switch {
	case v.increasing && int(v.value)+int(size) >= int(v.ceiling):
		v.increasing = false
		v.reversed = true
	case !v.increasing && int(v.value)-int(size) < int(v.floor):
		v.increasing = true
		v.reversed = true
	default:
		v.reversed = false
	}
	if v.increasing {
		v.value += size
	} else {
		v.value -= size
	}

	return v
}

func (v animatedValue) updateBounds(
	minFloor, maxFloor, minCeiling, maxCeiling uint8,
) animatedValue {
	if v.reversed {
		if v.increasing {
			//nolint:gosec // We don't need cryptographically-secure random number generation here
			v.floor = minFloor + uint8(rand.Intn(int(maxFloor)-int(minFloor)))
		} else {
			//nolint:gosec // We don't need cryptographically-secure random number generation here
			v.ceiling = minCeiling + uint8(rand.Intn(int(maxCeiling)-int(minCeiling)))
		}
	}
	return v
}

type animatedColor struct {
	r animatedValue
	g animatedValue
	b animatedValue
}

func newAnimatedColor(initial color.RGBA) animatedColor {
	const floor = 50
	const ceiling = 200
	return animatedColor{
		r: animatedValue{
			value:      initial.R,
			increasing: true,
			floor:      floor,
			ceiling:    ceiling,
		},
		g: animatedValue{
			value:      initial.G,
			increasing: true,
			floor:      floor,
			ceiling:    ceiling,
		},
		b: animatedValue{
			value:      initial.B,
			increasing: true,
			floor:      floor,
			ceiling:    ceiling,
		},
	}
}

func (c animatedColor) color() color.RGBA {
	const alpha = 255
	return color.RGBA{
		R: c.r.value,
		G: c.g.value,
		B: c.b.value,
		A: alpha,
	}
}

func (c animatedColor) step(size uint8) animatedColor {
	const minFloor = 50
	const maxFloor = 150
	const maxCeiling = 250
	c.r = c.r.step(size).updateBounds(minFloor, maxFloor, maxFloor+size, maxCeiling)
	c.g = c.g.step(size).updateBounds(minFloor, maxFloor, maxFloor+size, maxCeiling)
	c.b = c.b.step(size).updateBounds(minFloor, maxFloor, maxFloor+size, maxCeiling)
	return c
}

// Handlers

func (h *Handlers) HandleAnimatedColorFrameGet() echo.HandlerFunc {
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

		// Subscribe to source stream
		ctx, cancel := context.WithCancel(c.Request().Context())
		source := "/video-streams/animated-color/source.mjpeg"
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Generate data
		frame := <-frameBuffer
		cancel()
		f, err := frame.AsImageFrame()
		if err != nil {
			return errors.Wrapf(err, "error with animation source")
		}
		co := f.Im.At(f.Im.Bounds().Min.X, f.Im.Bounds().Min.Y)
		output := image.NewRGBA(image.Rect(0, 0, width, height))
		fillImage(output, co)
		f = &videostreams.ImageFrame{
			Im:   output,
			Meta: f.Meta.WithOp(videostreams.Operationf("resize to %dx%d", width, height)),
		}
		f.Meta.Settings.JPEGEncodeQuality = quality

		// Produce output
		frameJPEG, _, err := f.AsJPEG()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}
		return c.Blob(http.StatusOK, "image/jpeg", frameJPEG)
	}
}

func animatedColorStreamFrameSender(
	ss *mjpeg.StreamSender, width, height int, annotated bool, quality int,
	fpsCounter *ratecounter.RateCounter, fpsPeriod float32,
) handling.Consumer[videostreams.Frame] {
	return func(frame videostreams.Frame) (done bool, err error) {
		if err = frame.Error(); err != nil {
			return false, errors.Wrapf(err, "error with animation source")
		}

		// Generate output
		f, err := frame.AsImageFrame()
		if err != nil {
			return false, errors.Wrap(err, "couldn't read frame as image")
		}
		co := f.Im.At(f.Im.Bounds().Min.X, f.Im.Bounds().Min.Y)
		output := image.NewRGBA(image.Rect(0, 0, width, height))
		fillImage(output, co)
		f = &videostreams.ImageFrame{
			Im:   output,
			Meta: f.Meta.WithOp(videostreams.Operationf("resize to %dx%d", width, height)),
		}
		f.Meta.Settings.JPEGEncodeQuality = quality
		if annotated {
			fpsCounter.Incr(1)
			metadata := videostreams.AnnotationMetadata{
				FPSCount:  fpsCounter.Rate(),
				FPSPeriod: fpsPeriod,
			}.WithFrameData(f)
			f = f.WithAnnotation(metadata.String(), 1)
		}
		frame = f

		// Send output
		if err = handling.Except(
			ss.SendFrame(frame), context.Canceled, syscall.EPIPE,
		); err != nil {
			return false, errors.Wrap(err, "couldn't send mjpeg frame")
		}
		return false, nil
	}
}

func (h *Handlers) HandleAnimatedColorStreamGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		annotated := c.QueryParam("annotated") == "true"
		width, height, err := parseImageDimensions(c.QueryParam("width"), c.QueryParam("height"))
		if err != nil {
			return errors.Wrap(err, "couldn't parse image dimension query parameter")
		}
		const defaultQuality = 20
		quality, err := parseIntParam(c.QueryParam("quality"), "quality", defaultQuality)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image quality query parameter")
		}

		// Set up output stream
		ss := mjpeg.StartStream(c.Response().Writer)
		defer func() {
			if err := handling.Except(ss.Close(), syscall.EPIPE); err != nil {
				c.Logger().Error(errors.Wrap(err, "couldn't close stream"))
			}
		}()

		// Subscribe to source stream
		ctx := c.Request().Context()
		source := "/video-streams/animated-color/source.mjpeg"
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Process and deliver stream
		const fpsPeriod = 2
		fpsCounter := ratecounter.NewRateCounter(fpsPeriod * time.Second)
		if err := handling.Except(
			handling.Consume(ctx, frameBuffer, animatedColorStreamFrameSender(
				ss, width, height, annotated, quality, fpsCounter, fpsPeriod,
			)),
			context.Canceled,
		); err != nil {
			c.Logger().Error(errors.Wrap(err, "failed to send animated color stream"))
		}
		return nil
	}
}

func (h *Handlers) HandleAnimatedColorSourcePub() videostreams.HandlerFunc {
	return func(c *videostreams.Context) error {
		// Publish periodically
		const updateInterval = 50 * time.Millisecond
		color := newAnimatedColor(newRandomColor())
		return handling.Except(
			handling.Repeat(c.Context(), updateInterval, func() (done bool, err error) {
				// Generate output
				const width = 1
				const height = 1
				const step = 2
				color = color.step(step)
				frame := newCurrentFrame(newUniformImage(width, height, color.color()))

				// Send output
				c.Publish(frame)
				return false, nil
			}),
			context.Canceled,
		)
	}
}
