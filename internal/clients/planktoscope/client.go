// Package planktoscope provides a high-level client for control of planktoscopes
package planktoscope

import (
	"encoding/json"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
)

type Planktoscope struct {
	Pump Pump
}

type Client struct {
	Config Config
	Logger godest.Logger
	MQTT   mqtt.Client

	stateLock sync.Mutex
	pump      Pump
}

func NewClient(c Config, l godest.Logger) (client *Client, err error) {
	client = &Client{}
	client.Config = c
	client.Logger = l

	c.MQTT.SetOnConnectHandler(client.handleConnected)
	c.MQTT.SetConnectionLostHandler(client.handleConnectionLost)
	c.MQTT.SetReconnectingHandler(client.handleReconnecting)
	client.MQTT = mqtt.NewClient(&c.MQTT)
	return client, nil
}

func (c *Client) GetState() Planktoscope {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	return Planktoscope{
		Pump: c.pump,
	}
}

// MQTT

func (c *Client) EstablishConnection() error {
	token := c.MQTT.Connect()
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) handleConnected(cm mqtt.Client) {
	c.Logger.Infof("connected to MQTT broker %s", c.Config.Broker().String())
	// FIXME: we might not want to use QOS 1 everywhere (depends on which messages are idempotent)
	cm.Subscribe("#", 1, c.handleMessage)
}

func (c *Client) handleConnectionLost(_ mqtt.Client, err error) {
	c.stateLock.Lock()
	defer c.stateLock.Unlock()

	c.pump.StateKnown = false
	c.Logger.Warn(err)
	// TODO: notify clients that control has been lost
}

func (c *Client) handleReconnecting(_ mqtt.Client, _ *mqtt.ClientOptions) {
	c.Logger.Warn("reconnecting to MQTT broker...")
}

func (c *Client) handleMessage(_ mqtt.Client, m mqtt.Message) {
	broker := c.Config.Broker().String()
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
		c.Logger.Infof("%s/%s: %v", c.Config.Broker().String(), m.Topic(), payload)
	case "status/pump":
		if err := c.handlePumpStatusUpdate(topic, m.Payload()); err != nil {
			c.Logger.Errorf(errors.Wrapf(
				err, "%s/%s: invalid payload %s", broker, topic, rawPayload,
			).Error())
		}
	}
}
