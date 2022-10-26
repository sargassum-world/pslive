package videostreams

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/paulbellamy/ratecounter"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
	"golang.org/x/image/draw"

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

func annotate(
	frame videostreams.Frame, jpegQuality int, metadata videostreams.Metadata,
) (result videostreams.ImageFrame, err error) {
	im, err := frame.Image()
	if err != nil {
		return videostreams.ImageFrame{}, errors.Wrap(err, "couldn't convert frame to image")
	}
	metadata.Width = im.Bounds().Max.X
	metadata.Height = im.Bounds().Max.Y
	metadata.Timestamp = frame.Time()
	result = videostreams.ImageFrame{
		Timestamp:   frame.Time(),
		Data:        im,
		JPEGQuality: jpegQuality,
	}
	output := videostreams.CopyForAnnotation(im)
	videostreams.Annotate(output, result.Timestamp, metadata)
	result.Data = output
	return result, nil
}

func resize(im image.Image, height int) image.Image {
	aspectRatio := float32(im.Bounds().Max.X) / float32(im.Bounds().Max.Y)
	width := int(aspectRatio * float32(height))
	output := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(output, output.Rect, im, im.Bounds(), draw.Over, nil)
	return output
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
		if err = frame.Err(); err != nil {
			return errors.Wrapf(err, "error with animation source")
		}
		im, err := frame.Image()
		if err != nil {
			return errors.Wrap(err, "couldn't convert frame to image")
		}
		output := resize(im, height)
		frame = videostreams.ImageFrame{
			Timestamp:   frame.Time(),
			Data:        output,
			JPEGQuality: quality,
		}

		// Produce output
		frameJPEG, err := frame.JPEG()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}
		const base = 10
		c.Response().Header().Set("X-Timestamp", strconv.FormatInt(frame.Time().UnixMilli(), base))
		return c.Blob(http.StatusOK, "image/jpeg", frameJPEG)
	}
}

func externalSourceFrameSender(
	ss *mjpeg.StreamSender, annotated bool, quality int,
	fpsCounter *ratecounter.RateCounter, fpsPeriod float32,
) handling.Consumer[videostreams.Frame] {
	return func(frame videostreams.Frame) (done bool, err error) {
		if err = frame.Err(); err != nil {
			return false, errors.Wrapf(err, "error with video source")
		}

		// Generate output
		// Note: without annotation the frame passes through directly, potentially without any
		// JPEG decoding/encoding in the pipeline
		if annotated {
			fpsCounter.Incr(1)
			if frame, err = annotate(frame, quality, videostreams.Metadata{
				FPSCount:  fpsCounter.Rate(),
				FPSPeriod: fpsPeriod,
				Quality:   quality,
			}); err != nil {
				return false, errors.Wrap(err, "couldn't annotate frame")
			}
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

func (h *Handlers) HandleExternalSourcePub() videostreams.HandlerFunc {
	return func(c *videostreams.Context) error {
		// Parse params from topic
		query, err := c.Query()
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
		reader, closer, err := mjpeg.StartURLReader(ctx, h.hc, source)
		if err != nil {
			err = errors.Wrapf(err, "couldn't start mjpeg reader for %s", source)
			c.Publish(videostreams.NewErrorFrame(err))
			return err
		}
		defer func() {
			_ = closer.Close()
		}()

		// Read MJPEG stream parts
		return handling.Except(
			handling.Repeat(ctx, 0, func() (done bool, err error) {
				// Load data
				encodedFrame, err := mjpeg.ReadFrame(reader)
				if err != nil {
					return false, errors.Wrap(err, "couldn't read mjpeg frame")
				}
				if encodedFrame == (mjpeg.Frame{}) {
					return true, nil
				}

				// Publish data
				c.Publish(encodedFrame)
				return false, nil
			}),
			context.Canceled,
		)
	}
}