// Package videostreams provides a pub-sub system for running video stream processing pipelines
// on-demand.
package videostreams

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"time"

	"github.com/pkg/errors"
)

// Metadata

type Settings struct {
	JPEGEncodeQuality int
}

type Metadata struct {
	FromSource  map[string][]string
	ReceiveTime time.Time
	Operations  *OpChain
	Settings    Settings
}

func (m *Metadata) WithOp(op Operation) *Metadata {
	if op == Nop {
		return m
	}
	return &Metadata{
		FromSource:  m.FromSource,
		ReceiveTime: m.ReceiveTime,
		Operations:  m.Operations.With(op),
		Settings:    m.Settings,
	}
}

func (m *Metadata) WithSettings(settings Settings) *Metadata {
	return &Metadata{
		FromSource:  m.FromSource,
		ReceiveTime: m.ReceiveTime,
		Operations:  m.Operations,
		Settings:    settings,
	}
}

// Frame

type Frame interface {
	AsImageFrame() (*ImageFrame, error)
	AsJPEGFrame() (*JPEGFrame, error)
	Error() error
}

// ImageFrame

type ImageFrame struct {
	Im   image.Image
	Meta *Metadata
	Err  error
}

func NewErrorFrame(err error) Frame {
	return &ImageFrame{
		Err: err,
	}
}

func (f *ImageFrame) AsImageFrame() (*ImageFrame, error) {
	return f, errors.Wrap(f.Err, "stream error")
}

func (f *ImageFrame) AsJPEGFrame() (*JPEGFrame, error) {
	if f.Meta == nil {
		return nil, errors.Errorf("unspecified jpeg quality due to missing metadata")
	}
	quality := f.Meta.Settings.JPEGEncodeQuality
	if quality < 1 || quality > 100 {
		return nil, errors.Errorf("invalid jpeg quality %d", quality)
	}

	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, f.Im, &jpeg.Options{
		Quality: quality,
	}); err != nil {
		return nil, errors.Wrap(err, "couldn't jpeg-encode image")
	}
	return &JPEGFrame{
		Im:   buf.Bytes(),
		Meta: f.Meta.WithOp(Operationf("encode as JPEG with quality %d", quality)),
	}, nil
}

func (f *ImageFrame) AsBase64Frame() (*Base64Frame, error) {
	jpegFrame, err := f.AsJPEGFrame()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't jpeg-encode image for base64 encoding")
	}
	base64Frame, err := jpegFrame.AsBase64Frame()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't base64-encode jpeg-encoded image")
	}
	return base64Frame, nil
}

func (f *ImageFrame) Error() error {
	return f.Err
}

func (f *ImageFrame) WithResizeToHeight(height int) *ImageFrame {
	return &ImageFrame{
		Im:   ResizeToHeight(f.Im, height),
		Meta: f.Meta.WithOp(Operationf("resize to height %d", height)),
	}
}

func (f *ImageFrame) WithAnnotation(annotations string, lines int) *ImageFrame {
	output := AddAnnotationPadding(f.Im, lines, 0)
	AnnotateTop(output, annotations, lines)
	return &ImageFrame{
		Im:   output,
		Meta: f.Meta.WithOp(Operationf("annotate")),
	}
}

// JPEG Frame

type JPEGFrame struct {
	Im   []byte
	Meta *Metadata
	Err  error
}

func (f *JPEGFrame) AsImageFrame() (*ImageFrame, error) {
	im, err := jpeg.Decode(bytes.NewReader(f.Im))
	if err != nil {
		return nil, err
	}
	return &ImageFrame{
		Im:   im,
		Meta: f.Meta.WithOp("decode JPEG"),
		Err:  f.Err,
	}, nil
}

func (f *JPEGFrame) AsJPEGFrame() (*JPEGFrame, error) {
	return f, f.Err
}

func (f *JPEGFrame) AsBase64Frame() (*Base64Frame, error) {
	return &Base64Frame{
		Im:   base64.StdEncoding.EncodeToString(f.Im),
		Meta: f.Meta.WithOp("encode as base64"),
	}, nil
}

func (f *JPEGFrame) Error() error {
	return f.Err
}

// Base64Frame

type Base64Frame struct {
	Im   string
	Meta *Metadata
	Err  error
}

func (f *Base64Frame) AsImageFrame() (*ImageFrame, error) {
	jpegFrame, err := f.AsJPEGFrame()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't base64-decode image for jpeg decoding")
	}
	imageFrame, err := jpegFrame.AsImageFrame()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't jpeg-decode base64-decoded image")
	}
	return imageFrame, nil
}

func (f *Base64Frame) AsJPEGFrame() (*JPEGFrame, error) {
	decoded, err := base64.StdEncoding.DecodeString(f.Im)
	if err != nil {
		return nil, errors.New("couldn't base64-decode image")
	}
	return &JPEGFrame{
		Im:   decoded,
		Meta: f.Meta.WithOp("decode as base64"),
	}, nil
}

func (f *Base64Frame) AsBase64Frame() (*Base64Frame, error) {
	return f, nil
}

func (f *Base64Frame) Error() error {
	return f.Err
}
