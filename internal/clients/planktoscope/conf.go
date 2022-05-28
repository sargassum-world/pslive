package planktoscope

import (
	"fmt"
	"time"

	"github.com/atrox/haikunatorgo"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "PLANKTOSCOPE_"

type Config struct {
	URL      string
	ClientID string
	MQTT     mqtt.ClientOptions
}

func GetConfig(brokerURL, clientInstanceID string) (c Config, err error) {
	c.URL = brokerURL

	client := env.GetString(envPrefix+"MQTT_CLIENT", "")
	if client == "" {
		client = haikunator.New().Haikunate()
	}
	if clientInstanceID == "" {
		clientInstanceID = haikunator.New().Haikunate()
	}
	c.ClientID = fmt.Sprintf("pslive/%s/ps/%s", client, clientInstanceID)

	c.MQTT, err = GetMQTTConfig(brokerURL, c.ClientID)
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make MQTT config")
	}

	return c, nil
}

func getMQTTConnectRetryInterval() (time.Duration, error) {
	const defaultInterval = 10 // default: 10 seconds
	intervalRaw, err := env.GetInt64(
		envPrefix+"MQTT_CONNECT_RETRY", defaultInterval,
	)
	if err != nil {
		return 0, err
	}
	return time.Duration(intervalRaw) * time.Second, nil
}

func getMQTTReconnectInterval() (time.Duration, error) {
	const defaultInterval = 10 // default: 10 seconds
	intervalRaw, err := env.GetInt64(
		envPrefix+"MQTT_RECONNECT", defaultInterval,
	)
	if err != nil {
		return 0, err
	}
	return time.Duration(intervalRaw) * time.Second, nil
}

func GetMQTTConfig(brokerURL, clientID string) (c mqtt.ClientOptions, err error) {
	if len(brokerURL) == 0 {
		// If no broker is provided, return a zero-valued config
		return mqtt.ClientOptions{}, nil
	}
	c.AddBroker(brokerURL)

	c.SetClientID(clientID)

	c.SetConnectRetry(true)
	connectRetryInterval, err := getMQTTConnectRetryInterval()
	if err != nil {
		return mqtt.ClientOptions{}, errors.Wrap(err, "couldn't make connect retry interval config")
	}
	c.SetConnectRetryInterval(connectRetryInterval)

	c.SetAutoReconnect(true)
	reconnectInterval, err := getMQTTReconnectInterval()
	if err != nil {
		return mqtt.ClientOptions{}, errors.Wrap(err, "couldn't make reconnect interval config")
	}
	c.SetMaxReconnectInterval(reconnectInterval)
	return c, nil
}
