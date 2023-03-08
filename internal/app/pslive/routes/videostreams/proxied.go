package videostreams

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/handling"

	"github.com/sargassum-world/pslive/internal/clients/mjpeg"
	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Helpers

func parseURLParam(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("missing query param 'url' to specify the external source")
	}
	source, err := url.QueryUnescape(raw)
	return source, errors.Wrapf(err, "couldn't parse source url %s", raw)
}

// Handlers

func (h *Handlers) HandleExternalSourceFrameGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		const defaultHeight = 360
		height, err := parseIntParam(c.QueryParam("height"), "height", defaultHeight)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image height query parameter")
		}
		const defaultQuality = 80
		quality, err := parseIntParam(c.QueryParam("quality"), "quality", defaultQuality)
		if err != nil {
			return errors.Wrap(err, "couldn't parse image quality query parameter")
		}
		rawURL := c.QueryParam("url")
		sourceURL, err := url.QueryUnescape(rawURL)
		if err != nil {
			return errors.Wrapf(err, "couldn't parse source url %s", rawURL)
		}

		// Subscribe to source stream
		ctx, cancel := context.WithCancel(c.Request().Context())
		source := fmt.Sprintf(
			"/video-streams/external-stream/source.mjpeg?url=%s", url.QueryEscape(sourceURL),
		)
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Generate data
		frame := <-frameBuffer
		cancel()
		if frame == nil {
			return errors.Wrap(err, "couldn't load frame")
		}
		f, err := frame.AsImageFrame()
		if err != nil {
			return errors.Wrap(err, "couldn't read frame as image")
		}
		f = f.WithResizeToHeight(height)
		f.Meta.Settings.JPEGEncodeQuality = quality
		frame = f

		// Produce output
		jpeg, err := frame.AsJPEGFrame()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}
		return c.Blob(http.StatusOK, "image/jpeg", jpeg.Im)
	}
}

func externalSourceFrameSender(
	ss *mjpeg.StreamSender, annotated bool, quality int,
	fpsCounter *ratecounter.RateCounter, fpsPeriod float32,
	l godest.Logger,
) handling.Consumer[videostreams.Frame] {
	return func(frame videostreams.Frame) (done bool, err error) {
		l.Debug("received frame from external source")
		if err = frame.Error(); err != nil {
			return false, errors.Wrapf(err, "error with video source")
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
		l.Debug("sending frame from external source")
		if err := handling.Except(
			ss.SendFrame(frame), context.Canceled, syscall.EPIPE,
		); err != nil {
			return false, errors.Wrap(err, "couldn't send mjpeg frame")
		}
		return false, nil
	}
}

func (h *Handlers) HandleExternalSourceStreamGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		annotated := c.QueryParam("annotated") == "true"
		rawURL := c.QueryParam("url")
		sourceURL, err := url.QueryUnescape(rawURL)
		if err != nil {
			return errors.Wrapf(err, "couldn't parse source url %s", rawURL)
		}
		// TODO: implement a max framerate

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
		c.Logger().Debugf("subscribing to external stream source %s", sourceURL)
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Post-process and deliver stream
		const fpsPeriod = 2
		fpsCounter := ratecounter.NewRateCounter(fpsPeriod * time.Second)
		const quality = 50 // TODO: implement adaptive quality for min FPS
		if err := handling.Except(
			handling.Consume(ctx, frameBuffer, externalSourceFrameSender(
				ss, annotated, quality, fpsCounter, fpsPeriod, c.Logger(),
			)),
			context.Canceled,
		); err != nil {
			c.Logger().Error(errors.Wrapf(err, "failed to proxy stream %s", sourceURL))
		}
		return nil
	}
}

func (h *Handlers) HandleExternalSourcePub() videostreams.HandlerFunc {
	return func(c *videostreams.Context) error {
		// Parse params from topic
		query, err := c.QueryParams()
		if err != nil {
			err = errors.Wrap(err, "couldn't parse topic query params")
			c.Publish(videostreams.NewErrorFrame(err))
			return err
		}
		source, err := parseURLParam(query.Get("url"))
		if err != nil {
			c.Publish(videostreams.NewErrorFrame(err))
			return err
		}

		// Start reading the MJPEG stream
		ctx := c.Context()
		r, err := mjpeg.NewReceiverFromURL(ctx, h.hc, source)
		if err != nil {
			err = errors.Wrapf(err, "couldn't start mjpeg receiver for %s", source)
			c.Publish(videostreams.NewErrorFrame(err))
			return err
		}
		defer r.Close()

		// Read MJPEG stream parts
		c.Logger().Debugf("receiving mjpeg frames from %s", source)
		return handling.Except(
			handling.Repeat(ctx, 0, func() (done bool, err error) {
				// Load data
				c.Logger().Debugf("receiving mjpeg frame from %s", source)
				encodedFrame, err := r.Receive()
				if errors.Is(err, io.EOF) {
					c.Logger().Debugf("received eof from %s", source)
					return true, nil
				}
				if err != nil {
					c.Logger().Errorf("couldn't receive mjpeg frame from %s: %s", source, err)
					return false, errors.Wrap(err, "couldn't read mjpeg frame")
				}

				// Publish data
				c.Logger().Debugf("received and publishing frame from %s", source)
				c.Publish(encodedFrame)
				return false, nil
			}),
			context.Canceled,
		)
	}
}
