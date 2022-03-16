// Package planktoscope provides a high-level client for control of planktoscopes
package planktoscope

import (
	"encoding/json"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
)

const (
	mqttAtLeastOnce = 1
	mqttExactlyOnce = 2
)

type Planktoscope struct {
	Pump         Pump
	PumpSettings PumpSettings
}

type Client struct {
	Config Config
	Logger godest.Logger
	MQTT   mqtt.Client

	stateL       *sync.RWMutex
	pump         Pump
	pumpB        *Broadcaster
	pumpSettings PumpSettings
}

func NewClient(c Config, l godest.Logger) (client *Client, err error) {
	client = &Client{}
	client.Config = c
	client.Logger = l
	client.stateL = &sync.RWMutex{}
	client.pumpB = NewBroadcaster()
	const defaultVolume = 1
	const defaultFlowrate = 0.1
	client.pumpSettings = PumpSettings{
		Forward:  true,
		Volume:   defaultVolume,
		Flowrate: defaultFlowrate,
	}

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
		Pump:         c.pump,
		PumpSettings: c.pumpSettings,
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
	// FIXME: we might not want to use Once 1 everywhere (depends on which messages are idempotent)
	token := cm.Subscribe("#", mqttAtLeastOnce, c.handleMessage)
	go func(t mqtt.Token) {
		if t.Wait(); t.Error() != nil {
			c.Logger.Error(errors.Wrap(t.Error(), "couldn't subscribe to #"))
		}
	}(token)
}

func (c *Client) handleConnectionLost(_ mqtt.Client, err error) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

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
