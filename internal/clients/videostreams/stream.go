// Package videostreams provides a pub-sub system for running video stream processing pipelines
// on-demand.
package videostreams

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"time"

	"github.com/pkg/errors"
)

type Frame interface {
	Time() time.Time
	Image() (image.Image, error)
	JPEG() ([]byte, error)
	Err() error
}

func NewErrorFrame(err error) Frame {
	return ImageFrame{
		Error: err,
	}
}

type ImageFrame struct {
	Timestamp   time.Time
	Data        image.Image
	JPEGQuality int
	Error       error
}

func (f ImageFrame) Time() time.Time {
	return f.Timestamp
}

func (f ImageFrame) JPEG() ([]byte, error) {
	if f.JPEGQuality < 1 || f.JPEGQuality > 100 {
		return nil, errors.Errorf("invalid jpeg quality %d", f.JPEGQuality)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, f.Data, &jpeg.Options{
		Quality: f.JPEGQuality,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't jpeg-encode image")
	}
	return buf.Bytes(), nil
}

func (f ImageFrame) Image() (image.Image, error) {
	return f.Data, nil
}

func (f ImageFrame) CopyRGBA() ImageFrame {
	im := image.NewRGBA(f.Data.Bounds())
	draw.Draw(im, im.Bounds(), f.Data, f.Data.Bounds().Min, draw.Src)
	return ImageFrame{
		Timestamp: f.Timestamp,
		Data:      im,
	}
}

func (f ImageFrame) Err() error {
	return f.Error
}
