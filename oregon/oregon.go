package oregon

import (
	"strings"
)

// Config for Oregon Scientific Sensor (localizer)
type Config struct {
	Channel  int    `json:"channel"`
	ID       int    `json:"id"`
	Floor    string `json:"floor"`
	Location string `json:"location"`
}

// Rtl433 are datas retrieved from MQTT (sent by RTL_433)
type Rtl433 struct {
	Time        string  `json:"time"`
	ID          int     `json:"id"`
	Channel     int     `json:"channel"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature_C"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure_hPa"`
	Battery     float64 `json:"battery_ok"`
}

// IsOregon tests if this is an oregon sensor
func (r Rtl433) IsOregon() bool {
	return strings.HasPrefix(r.Model, "Oregon")
}

// Contains checks if the Rtl433 ID and Channel is
// in one of the Config in the given slice of Config.
// If yes, returns the Config and true
func Contains(s []Config, r Rtl433) (*Config, bool) {
	for _, c := range s {
		if c.ID == r.ID && c.Channel == r.Channel {
			return &c, true
		}
	}
	return nil, false
}

// SensorBase are common base datas (always sent to MQTT)
type SensorBase struct {
	Time     string `json:"time"`
	ID       int    `json:"id"`
	Floor    string `json:"floor"`
	Location string `json:"location"`
}

// Measurements are all measurements sent by the Oregon sensor
// used to send datas to influxdb
type Measurements struct {
	*SensorBase
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity,omitempty"`
	Pressure    float64 `json:"pressure,omitempty"`
	Battery     float64 `json:"battery,omitempty"`
}
