// Package planktoscope provides a high-level client for control of planktoscopes
package planktoscope

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
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
	mqttConnMu           *sync.Mutex
	firstConnSuccess     chan struct{}
	firstConnSuccessOnce *sync.Once

	stateL       *sync.RWMutex
	pump         Pump
	pumpB        *Broadcaster
	pumpSettings PumpSettings
}

func NewClient(c Config, l godest.Logger) (client *Client, err error) {
	client = &Client{}
	client.Config = c
	client.Logger = l
	client.mqttConnMu = &sync.Mutex{}
	client.firstConnSuccess = make(chan struct{})
	client.firstConnSuccessOnce = &sync.Once{}
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

func (c *Client) handleConnected(cm mqtt.Client) {
	c.firstConnSuccessOnce.Do(func() {
		close(c.firstConnSuccess)
	})
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
	case "actuator/pump":
		if err := c.handlePumpActuatorUpdate(topic, m.Payload()); err != nil {
			c.Logger.Errorf(errors.Wrapf(
				err, "%s/%s: invalid payload %s", broker, topic, rawPayload,
			).Error())
		}
	}
}

func (c *Client) Connect() error {
	c.mqttConnMu.Lock()
	defer c.mqttConnMu.Unlock()

	token := c.MQTT.Connect()
	_ = token.Wait()
	return errors.Wrapf(token.Error(), "couldn't connect to %s", c.Config.URL)
}

func (c *Client) ConnectedAtLeastOnce() <-chan struct{} {
	return c.firstConnSuccess
}

func (c *Client) Shutdown(ctx context.Context) error {
	if !c.MQTT.IsConnected() {
		return nil
	}

	// If the connection never opens, the Connect() method will never release mqttConnMu, but there
	// also (as far as I can tell) won't be any data race between MQTT.Connect() and MQTT.Disconnect()
	// which would otherwise require mqttConnMu. The paho.mqtt.golang package should really try to
	// avoid data races between its Connect and Disconnect methods...
	select {
	default:
		break
	case <-c.firstConnSuccess:
		c.mqttConnMu.Lock()
		defer c.mqttConnMu.Unlock()
	}

	closedNormally := make(chan struct{})
	go func() {
		const closeTimeout = 5000 // ms
		c.MQTT.Disconnect(closeTimeout)
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
	fmt.Println(c.Config.URL, "acquiring close lock on mqtt connection...")
	c.mqttConnMu.Lock()
	defer c.mqttConnMu.Unlock()
	fmt.Println(c.Config.URL, "acquired close lock on mqtt connection!")

	c.MQTT.Disconnect(0)
}
