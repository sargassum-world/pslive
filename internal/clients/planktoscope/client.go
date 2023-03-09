// Package planktoscope provides a high-level client for control of planktoscopes
package planktoscope

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
)

const (
	Protocol        = "planktoscope-v2.3"
	mqttAtLeastOnce = 1
	mqttExactlyOnce = 2
)

type Client struct {
	Config               Config
	Logger               godest.Logger
	MQTT                 mqtt.Client
	firstConnSuccess     chan struct{}
	firstConnSuccessOnce *sync.Once
	logReconnectOnce     *sync.Once
	logReconnectOnceMu   *sync.Mutex

	stateL         *sync.RWMutex
	pump           Pump
	pumpB          *Broadcaster
	pumpSettings   PumpSettings
	cameraB        *Broadcaster
	cameraSettings CameraSettings
}

func NewClient(c Config, l godest.Logger) (client *Client, err error) {
	client = &Client{}
	client.Config = c
	client.Logger = l
	client.firstConnSuccess = make(chan struct{})
	client.firstConnSuccessOnce = &sync.Once{}
	client.logReconnectOnce = &sync.Once{}
	client.logReconnectOnceMu = &sync.Mutex{}
	client.stateL = &sync.RWMutex{}
	client.pumpB = NewBroadcaster()
	client.pumpSettings = DefaultPumpSettings()
	client.cameraB = NewBroadcaster()
	client.cameraSettings = DefaultCameraSettings()

	c.MQTT.SetOnConnectHandler(client.handleConnected)
	c.MQTT.SetConnectionLostHandler(client.handleConnectionLost)
	c.MQTT.SetReconnectingHandler(client.handleReconnecting)
	client.MQTT = mqtt.NewClient(&c.MQTT)
	return client, nil
}

func (c *Client) GetState() Planktoscope {
	c.stateL.RLock()
	defer c.stateL.RUnlock()

	return Planktoscope{
		Pump:           c.pump,
		PumpSettings:   c.pumpSettings,
		CameraSettings: c.cameraSettings,
	}
}

// MQTT

func (c *Client) handleConnected(cm mqtt.Client) {
	c.firstConnSuccessOnce.Do(func() {
		close(c.firstConnSuccess)
	})
	c.Logger.Infof("connected as %s to MQTT broker %s", c.Config.ClientID, c.Config.URL)
	// FIXME: we might not want to use Once 1 everywhere (depends on which messages are idempotent)
	token := cm.Subscribe("#", mqttAtLeastOnce, c.handleMessage)
	go func(t mqtt.Token) {
		if t.Wait(); t.Error() != nil {
			c.Logger.Error(errors.Wrap(t.Error(), "couldn't subscribe to #"))
		}
	}(token)

	c.logReconnectOnceMu.Lock()
	c.logReconnectOnce = &sync.Once{}
	c.logReconnectOnceMu.Unlock()
}

func (c *Client) handleConnectionLost(_ mqtt.Client, err error) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.pump.StateKnown = false
	c.cameraSettings.StateKnown = false
	c.Logger.Warn(errors.Wrap(err, "connection lost"))
	// TODO: notify clients that control has been lost
}

func (c *Client) handleReconnecting(_ mqtt.Client, _ *mqtt.ClientOptions) {
	c.logReconnectOnceMu.Lock()
	defer c.logReconnectOnceMu.Unlock()

	c.logReconnectOnce.Do(func() {
		c.Logger.Warn("reconnecting to MQTT broker...")
	})
}

func (c *Client) handleMessage(_ mqtt.Client, m mqtt.Message) {
	broker := c.Config.URL
	rawPayload := string(m.Payload())

	switch topic := m.Topic(); topic {
	default:
		var payload interface{}
		if err := json.Unmarshal(m.Payload(), &payload); err != nil {
			c.Logger.Errorf(
				"%s/%s: unparseable payload %s", broker, topic, rawPayload,
			)
			return
		}
		c.Logger.Infof("%s/%s: %v", broker, m.Topic(), payload)
	case "status/pump":
		if err := c.handlePumpStatusUpdate(topic, m.Payload()); err != nil {
			c.Logger.Errorf(errors.Wrapf(
				err, "%s/%s: invalid payload %s", broker, topic, rawPayload,
			).Error())
		}
	case "actuator/pump":
		if err := c.handlePumpActuatorUpdate(topic, m.Payload()); err != nil {
			c.Logger.Errorf(errors.Wrapf(
				err, "%s/%s: invalid payload %s", broker, topic, rawPayload,
			).Error())
		}
	case "imager/image":
		// Because the PlanktoScope API never doesn't report the camera settings in a status message,
		// we must instead listen for imager camera settings update commands
		if err := c.handleCameraSettingsUpdate(topic, m.Payload()); err != nil {
			c.Logger.Errorf(errors.Wrapf(
				err, "%s/%s: invalid payload %s", broker, topic, rawPayload,
			).Error())
		}
	}
}

func (c *Client) Connect() error {
	token := c.MQTT.Connect()
	_ = token.Wait()
	return errors.Wrapf(token.Error(), "couldn't connect to %s", c.Config.URL)
}

func (c *Client) ConnectedAtLeastOnce() <-chan struct{} {
	return c.firstConnSuccess
}

func (c *Client) HasConnection() bool {
	return c.MQTT.IsConnectionOpen()
}

func (c *Client) Shutdown(ctx context.Context) error {
	if !c.MQTT.IsConnected() {
		return nil
	}

	const defaultTimeout = 5000 // ms
	var timeout uint
	select {
	default:
		// As of v1.4.2 of paho.mqtt.golang, c.MQTT.Disconnect seems hang beyond our close timeout if
		// the client never successfully connected - this implies that c.MQTT.IsConnected returns true
		// even for such clients.
		timeout = 0
		break
	case <-c.firstConnSuccess:
		timeout = defaultTimeout
		break
	}

	closedNormally := make(chan struct{})
	go func() {
		c.MQTT.Disconnect(timeout)
		close(closedNormally)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-closedNormally:
		return nil
	}
}

func (c *Client) Close() {
	if !c.MQTT.IsConnected() {
		return
	}

	c.MQTT.Disconnect(0)
}
