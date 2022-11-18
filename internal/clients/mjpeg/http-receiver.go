// Package mjpeg provides functionality for receiving and sending MJPEG streams over HTTP
package mjpeg

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

// Receiver

type Receiver struct {
	reader *multipart.Reader
	closer func() error
}

func NewReceiver(r io.Reader, boundary string) *Receiver {
	return &Receiver{
		reader: multipart.NewReader(r, boundary),
	}
}

func NewReceiverFromResponse(res *http.Response) (*Receiver, error) {
	contentType, param, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse content type from http header")
	}
	if !strings.HasPrefix(contentType, "multipart/") {
		return nil, errors.Errorf("unexpected stream content type %s", contentType)
	}
	receiver := NewReceiver(res.Body, param["boundary"])
	receiver.closer = res.Body.Close
	return receiver, nil
}

func NewReceiverFromURL(
	ctx context.Context, client *http.Client, u string,
) (*Receiver, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't make http get request for %s", u)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't open response from %s", u)
	}
	if res.StatusCode != http.StatusOK {
		_ = res.Body.Close()
		return nil, errors.Errorf("received http %d from %s", res.StatusCode, u)
	}
	receiver, err := NewReceiverFromResponse(res)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't start response receiver")
	}
	return receiver, nil
}

func (r *Receiver) Close() {
	if r.closer != nil {
		_ = r.closer()
	}
}

func (r *Receiver) Receive() (frame *videostreams.JPEGFrame, err error) {
	part, err := r.reader.NextPart()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get next part from stream")
	}
	if contentType := part.Header.Get("Content-Type"); contentType != "image/jpeg" {
		return nil, errors.Errorf("unexpected stream part content type %s", contentType)
	}

	// Copy image data
	buffer := &bytes.Buffer{}
	if _, err = io.Copy(buffer, part); err != nil {
		return nil, errors.Wrap(err, "couldn't jpeg-decode stream part")
	}
	frame = &videostreams.JPEGFrame{
		Im: buffer.Bytes(),
		Meta: &videostreams.Metadata{
			FromSource:  make(map[string][]string),
			ReceiveTime: time.Now(),
			Operations: &videostreams.OpChain{
				Op: "receive MJPEG over HTTP",
			},
		},
	}

	// Copy header fields
	for key, values := range part.Header {
		if key == "Content-Type" {
			continue
		}
		frame.Meta.FromSource[key] = values
	}
	return frame, err
}
