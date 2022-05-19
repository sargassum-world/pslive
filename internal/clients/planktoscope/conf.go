package planktoscope

import (
	"net/url"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "PLANKTOSCOPE_"

type Config struct {
	URL  string
	MQTT mqtt.ClientOptions
}

func (c Config) Broker() *url.URL {
	for _, serverURL := range c.MQTT.Servers {
		if serverURL != nil {
			return serverURL
		}
	}
	return &url.URL{}
}

func GetConfig(brokerURL string) (c Config, err error) {
	c.URL = brokerURL
	c.MQTT, err = GetMQTTConfig(brokerURL)
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

func GetMQTTConfig(brokerURL string) (c mqtt.ClientOptions, err error) {
	if len(brokerURL) == 0 {
		// If no broker is provided, return a zero-valued config
		return mqtt.ClientOptions{}, nil
	}
	c.AddBroker(brokerURL)
	c.SetAutoReconnect(true)
	c.SetClientID(env.GetString(envPrefix+"MQTT_CLIENTID", "pslive"))
	c.SetConnectRetry(true)

	connectRetryInterval, err := getMQTTConnectRetryInterval()
	if err != nil {
		return mqtt.ClientOptions{}, errors.Wrap(err, "couldn't make connect retry interval config")
	}
	c.SetConnectRetryInterval(connectRetryInterval)
	reconnectInterval, err := getMQTTReconnectInterval()
	if err != nil {
		return mqtt.ClientOptions{}, errors.Wrap(err, "couldn't make reconnect interval config")
	}
	c.SetMaxReconnectInterval(reconnectInterval)
	return c, nil
}
