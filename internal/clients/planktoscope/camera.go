package planktoscope

import (
	"encoding/json"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
)

func (c *Client) CameraStateBroadcasted() <-chan struct{} {
	return c.cameraB.Broadcasted()
}

// Receive Updates

func (c *Client) updateCameraSettings(newSettings CameraSettings) {
	c.stateL.Lock()
	defer c.stateL.Unlock()

	if newSettings.ISO > 0 {
		c.cameraSettings.ISO = newSettings.ISO
	}
	if newSettings.ShutterSpeed > 0 {
		c.cameraSettings.ShutterSpeed = newSettings.ShutterSpeed
	}
	c.cameraSettings.StateKnown = c.cameraSettings.ISO > 0 && c.cameraSettings.ShutterSpeed > 0
	c.cameraB.BroadcastNext()
}

func (c *Client) handleCameraSettingsUpdate(_ string, rawPayload []byte) error {
	type CameraSettingsCommand struct {
		Action   string `json:"action"`
		Settings struct {
			ISO          uint64 `json:"iso,omitempty"`
			ShutterSpeed uint64 `json:"shutter_speed,omitempty"`
		} `json:"settings,omitempty"`
	}
	var payload CameraSettingsCommand
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	if payload.Action != "settings" {
		return nil
	}

	newSettings := CameraSettings{}
	// TODO: do we need to call parseUint on the payload, e.g. if the Node-RED dashboard sends
	// non-uint values?
	newSettings.ISO = payload.Settings.ISO
	newSettings.ShutterSpeed = payload.Settings.ShutterSpeed

	// Commit changes
	c.updateCameraSettings(newSettings)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newSettings)
	return nil
}

// Send Commands

func (c *Client) SetCamera(iso, shutterSpeed uint64) (mqtt.Token, error) {
	type Settings struct {
		ISO          uint64 `json:"iso"`
		ShutterSpeed uint64 `json:"shutter_speed"`
	}
	command := struct {
		Action   string   `json:"action"`
		Settings Settings `json:"settings"`
	}{
		Action: "settings",
		Settings: Settings{
			ISO:          iso,
			ShutterSpeed: shutterSpeed,
		},
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	c.stateL.Lock()
	defer c.stateL.Unlock()

	c.cameraSettings.StateKnown = true
	c.cameraSettings.ISO = iso
	c.cameraSettings.ShutterSpeed = shutterSpeed
	// TODO: push updated settings to clients

	token := c.MQTT.Publish("imager/image", mqttExactlyOnce, false, marshaled)
	return token, nil
}
