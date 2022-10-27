package instruments

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/mjpeg"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Camera stream processing helpers

func parseIntParam(raw, name string, defaultValue int) (parsed int, err error) {
	parsed = defaultValue
	if raw != "" {
		if parsed, err = strconv.Atoi(raw); err != nil {
			return 0, errors.Wrapf(err, "invalid %s parameter %s", name, raw)
		}
	}
	return parsed, err
}

// Handlers

func (h *Handlers) HandleInstrumentCameraPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"camera",
		func(ctx context.Context, componentID int64, url, protocol string) error {
			return h.is.UpdateCamera(ctx, instruments.Camera{
				ID:       componentID,
				URL:      url,
				Protocol: protocol,
			})
		},
		h.is.DeleteCamera,
	)
}

func (h *Handlers) HandleInstrumentCameraFrameGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		id, err := parseID(c.Param("cameraID"), "camera")
		if err != nil {
			return err
		}
		const defaultHeight = 400
		height, err := parseIntParam(c.QueryParam("height"), "height", defaultHeight)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image height query parameter")
		}
		const defaultQuality = 80
		quality, err := parseIntParam(c.QueryParam("quality"), "quality", defaultQuality)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image quality query parameter")
		}

		// Run queries
		camera, err := h.is.GetCamera(c.Request().Context(), id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("camera %d not found", id))
		}
		sourceURL := camera.URL

		// Subscribe to source stream
		ctx, cancel := context.WithCancel(c.Request().Context())
		source := fmt.Sprintf(
			"/video-streams/external-stream/source.mjpeg?url=%s", url.QueryEscape(sourceURL),
		)
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Generate data
		frame := <-frameBuffer
		cancel()
		f, err := frame.AsImageFrame()
		if err != nil {
			return errors.Wrap(err, "couldn't read frame as image")
		}
		f = f.WithResizeToHeight(height)
		f.Meta.Settings.JPEGEncodeQuality = quality
		frame = f

		// Produce output
		jpeg, _, err := frame.AsJPEG()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}
		return c.Blob(http.StatusOK, "image/jpeg", jpeg)
	}
}

func externalSourceFrameSender(
	ss *mjpeg.StreamSender, annotated bool, quality int,
	fpsCounter *ratecounter.RateCounter, fpsPeriod float32,
) handling.Consumer[videostreams.Frame] {
	return func(frame videostreams.Frame) (done bool, err error) {
		if err = frame.Error(); err != nil {
			return false, errors.Wrapf(err, "error with stream source")
		}

		// Generate output
		// Note: without annotation the frame passes through directly, potentially without any
		// JPEG decoding/encoding in the pipeline
		if annotated {
			fpsCounter.Incr(1)
			f, err := frame.AsImageFrame()
			if err != nil {
				return false, errors.Wrap(err, "couldn't read frame as image")
			}
			f.Meta = f.Meta.WithSettings(videostreams.Settings{
				JPEGEncodeQuality: quality,
			})
			metadata := videostreams.AnnotationMetadata{
				FPSCount:  fpsCounter.Rate(),
				FPSPeriod: fpsPeriod,
			}.WithFrameData(f)
			f = f.WithAnnotation(metadata.String(), 1)
			frame = f
		}
		// TODO: implement image resizing

		// Send output
		if err := handling.Except(
			ss.SendFrame(frame), context.Canceled, syscall.EPIPE,
		); err != nil {
			return false, errors.Wrap(err, "couldn't send mjpeg frame")
		}
		return false, nil
	}
}

func (h *Handlers) HandleInstrumentCameraStreamGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		id, err := parseID(c.Param("cameraID"), "camera")
		if err != nil {
			return err
		}
		annotated := c.QueryParam("annotated") == "true"
		// TODO: implement a max framerate

		// Run queries
		camera, err := h.is.GetCamera(c.Request().Context(), id)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("camera %d not found", id))
		}
		sourceURL := camera.URL

		// Set up output stream
		ss := mjpeg.StartStream(c.Response().Writer)
		defer func() {
			if err := handling.Except(ss.Close(), syscall.EPIPE); err != nil {
				c.Logger().Error(errors.Wrap(err, "couldn't close stream"))
			}
		}()

		// Subscribe to source stream
		ctx := c.Request().Context()
		source := fmt.Sprintf(
			"/video-streams/external-stream/source.mjpeg?url=%s", url.QueryEscape(sourceURL),
		)
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Post-process and deliver stream
		const fpsPeriod = 2
		fpsCounter := ratecounter.NewRateCounter(fpsPeriod * time.Second)
		const quality = 50 // TODO: implement adaptive quality for min FPS
		if err := handling.Except(
			handling.Consume(ctx, frameBuffer, externalSourceFrameSender(
				ss, annotated, quality, fpsCounter, fpsPeriod,
			)),
			context.Canceled,
		); err != nil {
			c.Logger().Error(errors.Wrapf(err, "failed to proxy stream %s", sourceURL))
		}
		return nil
	}
}
