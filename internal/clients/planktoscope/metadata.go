package planktoscope

import (
	"encoding/json"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

// Send Commands

func (c *Client) SetMetadata(sampleID string, acquisitionTime time.Time) (mqtt.Token, error) {
	type Metadata struct {
		SampleID             string `json:"sample_id"`
		SampleCollectionDate string `json:"object_date"`
		SampleCollectionTime string `json:"object_time"`
		AcquisitionID        string `json:"acq_id"`
	}
	command := struct {
		Action   string   `json:"action"`
		Metadata Metadata `json:"config"`
	}{
		Action: "update_config",
		Metadata: Metadata{
			SampleID:             sampleID,
			SampleCollectionDate: acquisitionTime.Format("2006-01-02"),
			SampleCollectionTime: acquisitionTime.Format("15:04:05"),
			AcquisitionID:        acquisitionTime.Format(time.RFC3339),
		},
	}
	marshaled, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	token := c.MQTT.Publish("imager/image", mqttExactlyOnce, false, marshaled)
	return token, nil
}
