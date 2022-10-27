// Package videostreams contains the route handlers related to streaming video delivery
package videostreams

import (
	"net"
	"net/http"
	"time"

	"github.com/sargassum-world/godest"

	"github.com/sargassum-world/pslive/internal/clients/videostreams"
)

type Handlers struct {
	vsb *videostreams.Broker
	hc  *http.Client // for reading external streams
}

func New(vsb *videostreams.Broker) *Handlers {
	const readBufferSize = 1 << 20 // 1 MiB
	const dialTimeout = 5 * time.Second
	return &Handlers{
		vsb: vsb,
		hc: &http.Client{
			Transport: &http.Transport{
				ReadBufferSize: readBufferSize,
				DialContext: (&net.Dialer{
					Timeout: dialTimeout,
				}).DialContext,
			},
		},
	}
}

func (h *Handlers) Register(er godest.EchoRouter, vsr videostreams.Router) {
	er.GET("/video-streams/random-color/frame.jpeg", h.HandleRandomColorFrameGet())
	er.GET("/video-streams/random-color/stream.mjpeg", h.HandleRandomColorStreamGet())
	er.GET("/video-streams/animated-color/frame.jpeg", h.HandleAnimatedColorFrameGet())
	er.GET("/video-streams/animated-color/stream.mjpeg", h.HandleAnimatedColorStreamGet())
	vsr.PUB("/video-streams/animated-color/source.mjpeg", h.HandleAnimatedColorSourcePub())
	er.GET("/video-streams/external-stream/frame.jpeg", h.HandleExternalSourceFrameGet())
	er.GET("/video-streams/external-stream/stream.mjpeg", h.HandleExternalSourceStreamGet())
	vsr.PUB("/video-streams/external-stream/source.mjpeg", h.HandleExternalSourcePub())
}
