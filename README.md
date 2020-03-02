# low_energy_sensor_localizer

For personal usage (WIP):
* I use this in conjonction with [low_energy_sensor](https://github.com/conurb/low_energy_sensor) and __latest__ [rtl_433](https://github.com/merbanan/rtl_433).
* rtl_433 is already embedded in the Docker image (final image size is about 14Mb)

What it does:
* listens MQTT messages from rtl_433
* re-send messages on a normalized topic

eg, message received from rtl_433 on `rtl_433` topic:

`{"time":"2019-08-06 10:58:30","brand":"OS","model":"Oregon-BHTR968","id":204,"channel":2,"battery_ok":1,"temperature_C":24.2,"humidity":56,"pressure_hPa":1015.0}`

low_energy_sensor_localizer will re-send a __retained__ message. With this [low_energy_sensor_localizer.conf](./rtl_433/low_energy_sensor_localizer.conf), it will give us:

On MQTT topic `home/sensors/climate/firstfloor/corridor`:  
message: `{"time":"10:58:30","id":204,"temperature":24.2,"humidity":56,"pressure":1015,"battery":"HIGH"}`

## Modify these files with your choices:

* [rtl_433/rtl_433.env](./rtl_433/rtl_433.env)
* [low_energy_sensor_localizer.conf](./rtl_433/low_energy_sensor_localizer.conf) (json format)

## Build

eg for Raspberry
```bash
env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -o rtl_433/low_energy_sensor_localizer
docker-compose up -d
```

Optionnaly, you can strip the binary to save room (env. 2Mb) 