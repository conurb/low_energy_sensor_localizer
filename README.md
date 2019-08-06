# low_energy_sensor_localizer

For personal usage (WIP):
* I use this in conjonction with [low_energy_sensor](https://github.com/conurb/low_energy_sensor) and __latest__ [rtl_433](https://github.com/merbanan/rtl_433).

What it does:
* listens MQTT messages from rtl_433
* parses the message and send one or many new messages on a dedicated MQTT topic (retained mode) for each measure with the localization of the sensor
* sends retrieved measurements to [InfluxDB 2.0](https://portal.influxdata.com/downloads/) (presently in alpha version)  

low_energy_sensor_localizer listens for MQTT messages sent by rtl_433. For example `rtl_433 -R 12 -M newmodel -F mqtt://localhost,events=/rf_sensors` will write datas from Oregon Scientific Sensor on the MQTT topic `rf_sensors`. low_energy_sensor_localizer will parse the message and send a message (or many) on its own topic (with retained mode).

eg, message from rtl_433 on `/rf_sensors`:

`{"time":"2019-08-06 10:58:30","brand":"OS","model":"Oregon-BHTR968","id":204,"channel":2,"battery_ok":1,"temperature_C":24.2,"humidity":56,"pressure_hPa":1015.0}`

low_energy_sensor_localizer will send (with this [configuration.json](./configuration.json)):

On MQTT topic `/sensors/climate/firstfloor/corridor/temperature`:  
message: `{"time":"10:58:30","id":204,"floor":"firstfloor","location":"corridor","temperature":24.2}`

On MQTT topic `/sensors/climate/firstfloor/corridor/humidity`:  
message: `{"time":"10:58:30","id":204,"floor":"firstfloor","location":"corridor","humidity":56}`

On MQTT topic `/sensors/climate/firstfloor/corridor/pressure`:  
message: `{"time":"10:58:30","id":204,"floor":"firstfloor","location":"corridor","pressure":1015}`

On MQTT topic `/sensors/climate/firstfloor/corridor/battery`:  
message: `{"time":"10:58:30","id":204,"floor":"firstfloor","location":"corridor","battery":"HIGH"}`

So, for example, `/sensors/climate/firstfloor/+/temperature` will give us all the temperatures in the rooms on the firsfloor (for fast checking with `mosquitto_sub` or any other usage) and with Influxdb (or [grafana](https://grafana.com/)), we can have graphics for any room or combinaisons of rooms (average by hour, etc...).

Same thing could certainly be accomplished with node-red but I don't like this kind of js tools and where is the fun in the DIY process ? On the other hand, this should be far more lightweight than node-red on a raspi.

## Build for Raspberry

```bash
env GOOS=linux GOARCH=arm GOARM=5 go build
```