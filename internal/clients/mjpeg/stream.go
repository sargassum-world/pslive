// Package mjpeg provides decoding and encoding of MJPEG streams
package mjpeg

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
)

type EncodableFrame interface {
	Time() time.Time
	JPEG() ([]byte, error)
	Err() error
}

type Frame struct {
	Timestamp time.Time
	Data      *bytes.Buffer
	Error     error
}

func (f Frame) Time() time.Time {
	return f.Timestamp
}

func (f Frame) JPEG() ([]byte, error) {
	return f.Data.Bytes(), nil
}

func (f Frame) Image() (image.Image, error) {
	return jpeg.Decode(f.Data)
}

func (f Frame) Err() error {
	return f.Error
}

// Stream receiving

func StartReader(r io.Reader, boundary string) *multipart.Reader {
	return multipart.NewReader(r, boundary)
}

func StartResponseReader(res *http.Response) (*multipart.Reader, error) {
	contentType, param, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse content type from http header")
	}
	if !strings.HasPrefix(contentType, "multipart/") {
		return nil, errors.Errorf("unexpected stream content type %s", contentType)
	}
	return StartReader(res.Body, param["boundary"]), nil
}

func StartURLReader(
	ctx context.Context, client *http.Client, u string,
) (*multipart.Reader, io.Closer, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't make http get request for %s", u)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "couldn't open response from %s", u)
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, nil, errors.Errorf("received http %d from %s", res.StatusCode, u)
	}
	reader, err := StartResponseReader(res)
	if err != nil {
		_ = res.Body.Close()
		return nil, nil, errors.Wrap(err, "couldn't start response reader")
	}
	return reader, res.Body, nil
}

func ReadFrame(m *multipart.Reader) (frame Frame, err error) {
	part, err := m.NextPart()
	if err == io.EOF {
		return Frame{}, nil
	}
	if err != nil {
		return Frame{}, errors.Wrap(err, "couldn't get next part from stream")
	}
	if contentType := part.Header.Get("Content-Type"); contentType != "image/jpeg" {
		return Frame{}, errors.Errorf("unexpected stream part content type %s", contentType)
	}
	frame = Frame{
		Timestamp: time.Now(),
		Data:      &bytes.Buffer{},
	}
	const base = 10
	const timestampSize = 64
	if parsedTimestamp, perr := strconv.ParseInt(
		part.Header.Get("X-Timestamp"), base, timestampSize,
	); perr == nil {
		frame.Timestamp = time.UnixMilli(parsedTimestamp)
	}
	if _, err = io.Copy(frame.Data, part); err != nil {
		return Frame{}, errors.Wrap(err, "couldn't jpeg-decode stream part")
	}
	return frame, err
}

// Stream sending

type StreamSender struct {
	w  http.ResponseWriter
	mw *multipart.Writer
}

func StartStream(w http.ResponseWriter) *StreamSender {
	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+mw.Boundary())
	return &StreamSender{
		w:  w,
		mw: mw,
	}
}

func (ss *StreamSender) SendFrame(frame EncodableFrame) error {
	data, err := frame.JPEG()
	if err != nil {
		return errors.Wrap(err, "couldn't jpeg-encode frame")
	}

	h := textproto.MIMEHeader{}
	h.Set("Content-Type", "image/jpeg")
	h.Set("Content-Length", strconv.Itoa(len(data)))
	const base = 10
	h.Set("X-Timestamp", strconv.FormatInt(frame.Time().UnixMilli(), base))
	pw, err := ss.mw.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := pw.Write(data); err != nil {
		return err
	}
	if f, ok := ss.w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func (ss *StreamSender) Close() error {
	return ss.mw.Close()
}

func SendStream(
	ctx context.Context, w http.ResponseWriter, frames <-chan EncodableFrame,
) (err error) {
	ss := StartStream(w)
	defer func() {
		cerr := ss.mw.Close()
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	return handling.Consume(ctx, frames, func(frame EncodableFrame) (bool, error) {
		return false, ss.SendFrame(frame)
	})
}
