package instruments

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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

func newErrorFrame(width, height int, message string) *videostreams.ImageFrame {
	const max = 255
	const margin = 10
	output := image.NewRGBA(image.Rect(0, 0, width, height))
	fillImage(output, color.RGBA{0, 0, 0, max})
	const divisor = 2
	videostreams.AddLabel(
		output, message, color.RGBA{max, max, max, max},
		margin, (height-videostreams.LineHeight)/divisor,
	)
	const quality = 80
	return &videostreams.ImageFrame{
		Im: output,
		Meta: &videostreams.Metadata{
			Settings: videostreams.Settings{
				JPEGEncodeQuality: quality,
			},
		},
	}
}

func fillImage(im draw.Image, co color.Color) {
	draw.Draw(im, im.Bounds(), &image.Uniform{co}, image.Point{}, draw.Src)
}

// Error images

const (
	errorWidth  = 320
	errorHeight = 240
)

func newErrorJPEG(width, height int, message string) []byte {
	output, _, err := newErrorFrame(
		width, height, message,
	).AsJPEG()
	if err != nil {
		panic(err)
	}
	return output
}

var (
	jpegError    = newErrorJPEG(errorWidth, errorHeight, "stream failed")
	frameError   = newErrorFrame(errorWidth, errorHeight, "stream failed")
	frameLoading = newErrorFrame(errorWidth, errorHeight, "loading stream...")
)

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
			return errors.Wrap(err, "couldn't parse camera ID path parameter")
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
		if frame == nil {
			return c.Blob(http.StatusNotFound, "image/jpeg", jpegError)
		}
		f, err := frame.AsImageFrame()
		if err != nil {
			return c.Blob(http.StatusNotFound, "image/jpeg", jpegError)
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
			if herr := handling.Except(
				ss.SendFrame(frameError), context.Canceled, syscall.EPIPE,
			); herr != nil {
				return false, errors.Wrap(err, "couldn't send mjpeg error frame")
			}
			return false, err
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
		if err := handling.Except(ss.SendFrame(frame), context.Canceled, syscall.EPIPE); err != nil {
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
		if err := handling.Except(
			ss.SendFrame(frameLoading), context.Canceled, syscall.EPIPE,
		); err != nil {
			return errors.Wrap(err, "couldn't send mjpeg loading frame")
		}

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
