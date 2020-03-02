#!/bin/sh

low_energy_sensor_localizer -c /config &
rtl_433 -R 12 -M newmodel \
        -F  "mqtt://${MQTT_ADDRESS}:${MQTT_PORT},user=${MQTT_USER},pass=${MQTT_PASSWORD},events=${RTL_433_MQTT_TOPIC}"
