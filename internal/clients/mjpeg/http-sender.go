package mjpeg

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"

	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Sender

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

func (ss *StreamSender) SendFrame(frame videostreams.Frame) error {
	jpeg, err := frame.AsJPEGFrame()
	if err != nil {
		return errors.Wrap(err, "couldn't jpeg-encode frame")
	}
	return ss.Send(jpeg.Im)
}

func (ss *StreamSender) Send(frame []byte) error {
	h := textproto.MIMEHeader{}
	h.Set("Content-Type", "image/jpeg")
	h.Set("Content-Length", strconv.Itoa(len(frame)))
	pw, err := ss.mw.CreatePart(h)
	if err != nil {
		return err
	}
	if _, err := pw.Write(frame); err != nil {
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

func SendFrameStream(
	ctx context.Context, w http.ResponseWriter, frames <-chan videostreams.Frame,
) (err error) {
	ss := StartStream(w)
	defer func() {
		cerr := ss.mw.Close()
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	return handling.Consume(ctx, frames, func(frame videostreams.Frame) (bool, error) {
		return false, ss.SendFrame(frame)
	})
}

func SendStream(
	ctx context.Context, w http.ResponseWriter, frames <-chan []byte,
) (err error) {
	ss := StartStream(w)
	defer func() {
		cerr := ss.mw.Close()
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	return handling.Consume(ctx, frames, func(frame []byte) (bool, error) {
		return false, ss.Send(frame)
	})
}
