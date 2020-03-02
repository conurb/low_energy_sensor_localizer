package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/conurb/low_energy_sensor_localizer/oregon"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {

	var configPath = flag.String("c", "./", "path to the \"low_energy_sensor_localizer.conf\" file")
	flag.Parse()

	c, err := loadConfig(filepath.Join(*configPath, "low_energy_sensor_localizer.conf"))
	if err != nil {
		log.Printf("Error while parsing configuration file: %v\n", err)
		os.Exit(1)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	opts := MQTT.NewClientOptions().AddBroker(net.JoinHostPort(os.Getenv("MQTT_ADDRESS"), os.Getenv("MQTT_PORT")))
	opts.SetClientID("domotic-low_energy_sensor_localizer")
	opts.SetDefaultPublishHandler(handleWithConfig(&c))
	opts.SetUsername(os.Getenv("MQTT_USER"))
	opts.SetPassword(os.Getenv("MQTT_PASSWORD"))

	opts.OnConnect = func(mc MQTT.Client) {
		if token := mc.Subscribe(os.Getenv("RTL_433_MQTT_TOPIC"), 0, handleWithConfig(&c)); token.Wait() && token.Error() != nil {
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
	Oregons       []oregon.Config `json:"oregon"`
	MQTTTopicBase string          `json:"mqtt_send_topic_base"`
}

func makeTopicPath(path ...string) string {
	var s strings.Builder
	for _, str := range path {
		s.WriteString(str)
		s.WriteByte('/')
	}
	return strings.TrimSuffix(s.String(), "/")
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

			// MQTT TOPIC PATH
			topicPath := makeTopicPath(c.MQTTTopicBase, conf.Floor, conf.Location)

			// make oregon.Sensor
			os := oregon.Sensor{
				Time:        strings.Fields(rtl433.Time)[1],
				ID:          conf.ID,
				Temperature: rtl433.Temperature,
				Humidity:    rtl433.Humidity,
				Pressure:    rtl433.Pressure,
			}

			// MQTT: battery
			if rtl433.Battery != -1 {
				var battery string
				if rtl433.Battery == 1 {
					battery = "HIGH"
				} else if rtl433.Battery == 0 {
					battery = "LOW"
				} else {
					battery = "UNKNOWN"
				}
				os.Battery = battery
			}

			if measureJSON, err := json.Marshal(os); err == nil {
				token := client.Publish(topicPath, 0, true, measureJSON)
				token.Wait()
			}
		}
	}
	return mh
}
