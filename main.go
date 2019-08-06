package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/conurb/low_energy_sensor_localizer/oregon"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/influxdata/influxdb-client-go"
)

func main() {

	var influxToken = flag.String("t", "", "influxdb token (default read from env var INFLUX_TOKEN)")
	var configPath = flag.String("c", "./", "path to the \"configuration.json\" file")
	flag.Parse()

	c, err := loadConfig(filepath.Join(*configPath, "configuration.json"))
	if err != nil {
		log.Printf("Error while parsing configuration file: %v\n", err)
		os.Exit(1)
	}

	// INFLUX_TOKEN
	if *influxToken == "" {
		*influxToken = os.Getenv("INFLUX_TOKEN")
		if *influxToken == "" {
			fmt.Println("[ERROR] Influx token not found")
			os.Exit(1)
		}
	}
	c.InfluxToken = *influxToken

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	opts := MQTT.NewClientOptions().AddBroker(c.MQTTServer)
	opts.SetClientID("DFS-low_energy_sensor_localizer")
	opts.SetDefaultPublishHandler(handleWithConfig(&c))

	opts.OnConnect = func(mc MQTT.Client) {
		if token := mc.Subscribe(c.Rtl433Topic, 0, handleWithConfig(&c)); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to server")
	}
	<-ch
}

func loadConfig(filename string) (Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var c Config
	err = json.Unmarshal(b, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}

// Config contains datas retrieved from configuration.json
type Config struct {
	Oregons            []oregon.Config `json:"oregon"`
	MQTTServer         string          `json:"mqtt_server"`
	Rtl433Topic        string          `json:"rtl_433_topic"`
	MQTTTopicBase      string          `json:"mqtt_send_topic_base"`
	InfluxServer       string          `json:"influx_server"`
	InfluxOrganization string          `json:"influx_organization"`
	InfluxBucket       string          `json:"influx_bucket"`
	InfluxMeasurement  string          `json:"influx_measurement"`
	InfluxToken        string
}

func makeTopicPath(path ...string) string {
	var s strings.Builder
	for _, str := range path {
		s.WriteString(str)
		s.WriteByte('/')
	}
	return s.String()
}

func handleWithConfig(c *Config) MQTT.MessageHandler {
	var mh MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

		if len(c.Oregons) == 0 {
			return
		}

		var rtl433 oregon.Rtl433
		rtl433.Battery = -1

		err := json.NewDecoder(bytes.NewReader(msg.Payload())).Decode(&rtl433)
		if err != nil {
			log.Printf("error while parsing rtl_433 json: %v\n", err)
		}

		if !rtl433.IsOregon() {
			return
		}

		// if we own this sensor
		if conf, ok := oregon.Contains(c.Oregons, rtl433); ok {
			// make oregon.SensorBase
			osb := oregon.SensorBase{
				Time:     strings.Fields(rtl433.Time)[1],
				ID:       conf.ID,
				Floor:    conf.Floor,
				Location: conf.Location,
			}

			// INFLUXDB: store values in db
			sendToInflux(c, &oregon.Measurements{
				SensorBase:  &osb,
				Temperature: rtl433.Temperature,
				Humidity:    rtl433.Humidity,
				Pressure:    rtl433.Pressure,
				Battery:     rtl433.Battery,
			})

			// MQTT TOPIC PATH
			topicPath := makeTopicPath(c.MQTTTopicBase, conf.Floor, conf.Location)

			// MQTT: send temperature
			t := struct {
				*oregon.SensorBase
				Temperature float64 `json:"temperature"`
			}{
				SensorBase:  &osb,
				Temperature: rtl433.Temperature,
			}
			if measureJSON, err := json.Marshal(t); err == nil {
				token := client.Publish(topicPath+"temperature", 0, true, measureJSON)
				token.Wait()
			}

			// MQTT: send humidity
			if rtl433.Humidity != 0 {
				h := struct {
					*oregon.SensorBase
					Humidity float64 `json:"humidity"`
				}{
					SensorBase: &osb,
					Humidity:   rtl433.Humidity,
				}
				if measureJSON, err := json.Marshal(h); err == nil {
					token := client.Publish(topicPath+"humidity", 0, true, measureJSON)
					token.Wait()
				}
			}

			// MQTT: send pressure
			if rtl433.Pressure != 0 {
				p := struct {
					*oregon.SensorBase
					Pressure float64 `json:"pressure"`
				}{
					SensorBase: &osb,
					Pressure:   rtl433.Pressure,
				}
				if measureJSON, err := json.Marshal(p); err == nil {
					token := client.Publish(topicPath+"pressure", 0, true, measureJSON)
					token.Wait()
				}
			}

			// MQTT: send battery
			if rtl433.Battery != -1 {
				var battery string
				if rtl433.Battery == 1 {
					battery = "HIGH"
				} else if rtl433.Battery == 0 {
					battery = "LOW"
				} else {
					battery = "UNKNOWN"
				}
				b := struct {
					*oregon.SensorBase
					Battery string `json:"battery"`
				}{
					SensorBase: &osb,
					Battery:    battery,
				}
				if measureJSON, err := json.Marshal(b); err == nil {
					token := client.Publish(topicPath+"battery", 0, true, measureJSON)
					token.Wait()
				}
			}

		}
	}
	return mh
}

func sendToInflux(c *Config, m *oregon.Measurements) {
	tags := map[string]string{"floor": m.SensorBase.Floor, "location": m.SensorBase.Location}
	fields := make(map[string]interface{})

	// sensor always sends temperature
	fields["temperature"] = m.Temperature
	// sensor always sends battery state
	fields["battery"] = m.Battery

	if m.Humidity != 0 {
		fields["humidity"] = m.Humidity
	}
	if m.Pressure != 0 {
		fields["pressure"] = m.Pressure
	}

	influx, err := influxdb.New(c.InfluxServer, c.InfluxToken, influxdb.WithHTTPClient(&http.Client{}))
	if err != nil {
		log.Printf("[INFLUXDB] Failed while creating client: %v\n", err)
		return
	}

	defer influx.Close()

	metrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			fields,
			c.InfluxMeasurement,
			tags,
			time.Now().UTC()),
	}

	if err := influx.Write(context.Background(), c.InfluxBucket, c.InfluxOrganization, metrics...); err != nil {
		log.Printf("[INFLUXFB] Failed while writing datas: %v\n", err)
	}
}
