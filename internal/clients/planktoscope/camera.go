package planktoscope

import (
	"encoding/json"
	"math"

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
	c.cameraSettings.AutoWhiteBalance = newSettings.AutoWhiteBalance
	if newSettings.WhiteBalanceRedGain > 0 {
		c.cameraSettings.WhiteBalanceRedGain = newSettings.WhiteBalanceRedGain
	}
	if newSettings.WhiteBalanceBlueGain > 0 {
		c.cameraSettings.WhiteBalanceBlueGain = newSettings.WhiteBalanceBlueGain
	}
	c.cameraSettings.StateKnown = c.cameraSettings.ISO > 0 && c.cameraSettings.ShutterSpeed > 0 &&
		c.cameraSettings.WhiteBalanceRedGain > 0 && c.cameraSettings.WhiteBalanceBlueGain > 0
	c.cameraB.BroadcastNext()
}

func (c *Client) handleCameraSettingsUpdate(_ string, rawPayload []byte) error {
	type CameraSettingsCommand struct {
		Action   string `json:"action"`
		Settings struct {
			ISO              uint64 `json:"iso,omitempty"`
			ShutterSpeed     uint64 `json:"shutter_speed,omitempty"`
			WhiteBalance     string `json:"white_balance,omitempty"`
			WhiteBalanceGain struct {
				Red  float64 `json:"red,omitempty"`
				Blue float64 `json:"blue,omitempty"`
			} `json:"white_balance_gain,omitempty"`
		} `json:"settings,omitempty"`
	}
	var payload CameraSettingsCommand
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return errors.Wrapf(err, "unparseable payload")
	}
	if action := payload.Action; action != "settings" {
		return errors.Errorf("unknown action %s", action)
	}

	newSettings := CameraSettings{}
	newSettings.ISO = payload.Settings.ISO
	newSettings.ShutterSpeed = payload.Settings.ShutterSpeed
	newSettings.AutoWhiteBalance = payload.Settings.WhiteBalance == "auto"
	const whiteBalanceMultiplier = 100
	newSettings.WhiteBalanceRedGain = payload.Settings.WhiteBalanceGain.Red / whiteBalanceMultiplier
	newSettings.WhiteBalanceBlueGain = payload.Settings.WhiteBalanceGain.Blue / whiteBalanceMultiplier

	// Commit changes
	c.updateCameraSettings(newSettings)
	c.Logger.Debugf("%s: %+v", c.Config.URL, newSettings)
	return nil
}

// Send Commands

func (c *Client) SetCamera(
	iso, shutterSpeed uint64,
	autoWhiteBalance bool, whiteBalanceRedGain, whiteBalanceBlueGain float64,
) (mqtt.Token, error) {
	type WhiteBalanceGain struct {
		Red  float64 `json:"red,omitempty"`
		Blue float64 `json:"blue,omitempty"`
	}
	whiteBalance := "off"
	const whiteBalanceMultiplier = 100
	whiteBalanceGain := &WhiteBalanceGain{
		Red:  math.Round(whiteBalanceRedGain * whiteBalanceMultiplier),
		Blue: math.Round(whiteBalanceBlueGain * whiteBalanceMultiplier),
	}
	if autoWhiteBalance {
		whiteBalance = "auto"
		whiteBalanceGain = nil
	}

	type Settings struct {
		ISO          uint64 `json:"iso"`
		ShutterSpeed uint64 `json:"shutter_speed"`
		WhiteBalance string `json:"white_balance"`
		// If the gains are provided even with auto white balance, the backend reverts to manual
		// white balance behavior
		WhiteBalanceGain *WhiteBalanceGain `json:"white_balance_gain,omitempty"`
	}
	command := struct {
		Action   string   `json:"action"`
		Settings Settings `json:"settings"`
	}{
		Action: "settings",
		Settings: Settings{
			ISO:              iso,
			ShutterSpeed:     shutterSpeed,
			WhiteBalance:     whiteBalance,
			WhiteBalanceGain: whiteBalanceGain,
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
	c.cameraSettings.AutoWhiteBalance = autoWhiteBalance
	c.cameraSettings.WhiteBalanceRedGain = whiteBalanceRedGain
	c.cameraSettings.WhiteBalanceBlueGain = whiteBalanceBlueGain

	token := c.MQTT.Publish("imager/image", mqttExactlyOnce, false, marshaled)
	return token, nil
}
