package videostreams

import (
	"context"
	"sync"

	"github.com/sargassum-world/godest/pubsub"
)

// Context

type brokerContext = pubsub.BrokerContext[*Context, Frame]

type Context struct {
	*brokerContext
}

// Handlers

type (
	HandlerFunc    = pubsub.HandlerFunc[*Context]
	MiddlewareFunc = pubsub.MiddlewareFunc[*Context]
)

func EmptyHandler(c *Context) error {
	return nil
}

// Broker

type (
	Hub   = pubsub.Hub[[]Frame]
	Route = pubsub.Route
)

const (
	MethodPub   = pubsub.MethodPub
	MethodSub   = pubsub.MethodSub
	MethodUnsub = pubsub.MethodUnsub
)

type Broker struct {
	broker *pubsub.Broker[*Context, Frame]
	logger pubsub.Logger
}

func NewBroker(logger pubsub.Logger) *Broker {
	return &Broker{
		broker: pubsub.NewBroker[*Context, Frame](logger),
		logger: logger,
	}
}

func (b *Broker) Hub() *Hub {
	return b.broker.Hub()
}

func (b *Broker) PUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.broker.Add(pubsub.MethodPub, topic, h, m...)
}

func (b *Broker) SUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.broker.Add(pubsub.MethodSub, topic, h, m...)
}

func (b *Broker) UNSUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.broker.Add(pubsub.MethodUnsub, topic, h, m...)
}

func (b *Broker) Use(middleware ...MiddlewareFunc) {
	b.broker.Use(middleware...)
}

type BroadcastReceiver func(ctx context.Context, frames []Frame) (ok bool)

func (b *Broker) subscribe(
	ctx context.Context, topic string, broadcastHandler BroadcastReceiver,
) (unsubscriber func(), finished <-chan struct{}) {
	// we keep this private because we're handling frames and we don't want a slow consumer to block
	// every other consumer; the public Subscribe method drops frames for busy consumers
	return b.broker.Subscribe(
		ctx, topic, func(c *brokerContext) *Context {
			return &Context{
				brokerContext: c,
			}
		},
		broadcastHandler,
	)
}

func (b *Broker) Subscribe(ctx context.Context, topic string) <-chan Frame {
	buffer := make(chan Frame, 1)
	wg := sync.WaitGroup{}
	unsubscribe, finished := b.subscribe(
		ctx, topic, func(ctx context.Context, frames []Frame) (ok bool) {
			wg.Add(1)
			select {
			case buffer <- frames[len(frames)-1]:
			default:
				// The buffer consumer can't keep up, so drop the frame to avoid blocking other subscribers
				// TODO: update a frames-dropped-per-sec counter
			}
			wg.Done()
			return true
		},
	)
	go func() {
		select {
		case <-ctx.Done():
		case <-finished:
		}
		unsubscribe()
		wg.Wait() // prevent closing channel with pending sends, which is a data race
		close(buffer)
	}()
	return buffer
}

func (b *Broker) Serve(ctx context.Context) error {
	return b.broker.Serve(ctx, func(c *brokerContext) *Context {
		return &Context{
			brokerContext: c,
		}
	})
}

type Router = pubsub.Router[*Context]
