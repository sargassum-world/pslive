package videostreams

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/handling"
	"github.com/sargassum-world/godest/pubsub"
)

// ChannelName is the name of the Action Cable channel for Video Streams.
const ChannelName = "Video::StreamsChannel"

// subscriber creates a subscription for the channel, to integrate [Channel] with [Broker].
type subscriber func(ctx context.Context, topic string) <-chan Frame

// Channel represents an Action Cable channel for a Video Streams stream.
type Channel struct {
	identifier string
	streamName string
	h          *pubsub.Hub[[]Frame]
	subscriber subscriber
	sessionID  string
	logger     pubsub.Logger
}

// parseStreamName parses the Video Streams stream name from the Action Cable subscription
// identifier.
func parseStreamName(identifier string) (string, error) {
	var i struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(identifier), &i); err != nil {
		return "", errors.Wrap(
			err, "couldn't parse stream name from action cable subscription identifier",
		)
	}
	return i.Name, nil
}

// NewChannel checks the identifier with the specified checkers and returns a new Channel instance.
func NewChannel(
	identifier string, h *pubsub.Hub[[]Frame], subscriber subscriber, sessionID string,
	checkers []actioncable.IdentifierChecker, logger pubsub.Logger,
) (*Channel, error) {
	name, err := parseStreamName(identifier)
	if err != nil {
		return nil, err
	}
	for _, checker := range checkers {
		if err := checker(identifier); err != nil {
			return nil, errors.Wrap(err, "action cable subscription identifier failed checks")
		}
	}
	return &Channel{
		identifier: identifier,
		streamName: name,
		h:          h,
		subscriber: subscriber,
		sessionID:  sessionID,
		logger:     logger,
	}, nil
}

// Subscribe handles an Action Cable subscribe command from the client with the provided
// [actioncable.Subscription].
func (c *Channel) Subscribe(ctx context.Context, sub *actioncable.Subscription) error {
	if sub.Identifier() != c.identifier {
		return errors.Errorf(
			"channel identifier %+v does not match subscription identifier %+v",
			c.identifier, sub.Identifier(),
		)
	}

	frames := c.subscriber(ctx, c.streamName)
	go func() {
		for frame := range frames {
			encoded, err := frame.AsJPEGFrame()
			if err != nil {
				c.logger.Error(errors.Wrap(err, "couldn't jpeg-encode frame to send over action cable"))
				break
			}
			// TODO: attach metadata to the frame data (may need struct tags for marshaling)
			if err := handling.Except(sub.SendBytes(ctx, encoded.Im), context.Canceled); err != nil {
				c.logger.Error(errors.Wrap(err, "couldn't send turbo streams messages over action cable"))
				break
			}
		}
		sub.Close()
	}()
	return nil
}

// Perform handles an Action Cable action command from the client.
func (c *Channel) Perform(data string) error {
	return errors.New("video streams channel cannot perform any actions")
}

// NewChannelFactory creates an [actioncable.ChannelFactory] for Turbo Streams to create channels
// for different Video Streams streams as needed.
func NewChannelFactory(
	b *Broker, sessionID string, logger pubsub.Logger, checkers ...actioncable.IdentifierChecker,
) actioncable.ChannelFactory {
	return func(identifier string) (actioncable.Channel, error) {
		return NewChannel(identifier, b.Hub(), b.Subscribe, sessionID, checkers, logger)
	}
}
