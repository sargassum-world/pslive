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
	frame, err := newErrorFrame(
		width, height, message,
	).AsJPEGFrame()
	if err != nil {
		panic(err)
	}
	return frame.Im
}

var (
	frameError = newErrorFrame(errorWidth, errorHeight, "stream failed")
	jpegError  = newErrorJPEG(errorWidth, errorHeight, "stream failed")
)

// Sending helpers

func externalSourceFrameSender(
	ss *mjpeg.StreamSender, annotated bool, quality int,
	fpsCounter *ratecounter.RateCounter, fpsPeriod float32,
) handling.Consumer[videostreams.Frame] {
	return func(frame videostreams.Frame) (done bool, err error) {
		if err = frame.Error(); err != nil {
			if herr := handling.Except(
				ss.Send(jpegError), context.Canceled, syscall.EPIPE,
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

// Handlers

func (h *Handlers) HandleInstrumentCameraPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"camera",
		func(
			ctx context.Context, componentID instruments.CameraID, url, protocol string, enabled bool,
		) error {
			return h.is.UpdateCamera(ctx, instruments.Camera{
				ID:       componentID,
				URL:      url,
				Protocol: protocol,
				Enabled:  enabled,
			})
		},
		h.is.DeleteCamera,
	)
}

func (h *Handlers) HandleInstrumentCameraFrameGet() echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		// Parse params
		cameraID, err := parseID[instruments.CameraID](c.Param("cameraID"), "camera")
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
		camera, err := h.is.GetCamera(c.Request().Context(), cameraID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("camera %d not found", cameraID))
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
		jpeg, err := frame.AsJPEGFrame()
		if err != nil {
			return errors.Wrap(err, "couldn't jpeg-encode image")
		}
		return c.Blob(http.StatusOK, "image/jpeg", jpeg.Im)
	}
}

func (h *Handlers) HandleInstrumentCameraStreamGet() echo.HandlerFunc {
	jpegLoading := newErrorJPEG(errorWidth, errorHeight, "loading stream...")
	return func(c echo.Context) error {
		// Parse params
		cameraID, err := parseID[instruments.CameraID](c.Param("cameraID"), "camera")
		if err != nil {
			return err
		}
		annotated := c.QueryParam("annotated") == flagChecked
		// TODO: implement a max framerate

		// Run queries
		camera, err := h.is.GetCamera(c.Request().Context(), cameraID)
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("camera %d not found", cameraID))
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
			ss.Send(jpegLoading), context.Canceled, syscall.EPIPE,
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

func (h *Handlers) HandleInstrumentCameraStreamPub() videostreams.HandlerFunc {
	frameLoading := newErrorFrame(errorWidth, errorHeight, "loading stream...")
	return func(c *videostreams.Context) error {
		// Parse params
		cameraID, err := parseID[instruments.CameraID](c.Param("cameraID"), "camera")
		if err != nil {
			return err
		}
		// TODO: implement a max framerate

		// Run queries
		ctx := c.Context()
		camera, err := h.is.GetCamera(ctx, cameraID)
		if err != nil {
			return errors.Wrapf(err, "camera %d not found", cameraID)
		}
		sourceURL := camera.URL

		// Set up output stream
		c.Publish(frameLoading)

		// Subscribe to source stream
		source := fmt.Sprintf(
			"/video-streams/external-stream/source.mjpeg?url=%s", url.QueryEscape(sourceURL),
		)
		frameBuffer := h.vsb.Subscribe(ctx, source)

		// Post-process and deliver stream
		if err := handling.Except(
			handling.Consume(ctx, frameBuffer, func(frame videostreams.Frame) (done bool, err error) {
				// Since the source stream emits JPEG frames, we can safely assume that no further JPEG
				// encoding is needed after we call c.Publish - but we check here anyways because it's
				// cheap to check, and we can ensure that we only perform one JPEG encoding regardless of
				// the number of Video Streams subscribers (i.e. the number of web browsers).
				jpegFrame, err := frame.AsJPEGFrame()
				if err != nil {
					return false, errors.Wrap(err, "couldn't convert frame to JPEG for action cable")
				}
				c.Publish(jpegFrame)
				return false, nil
			}),
			context.Canceled,
		); err != nil {
			c.Publish(frameError)
			c.Logger().Error(errors.Wrapf(err, "failed to proxy stream %s", sourceURL))
		}
		return nil
	}
}
