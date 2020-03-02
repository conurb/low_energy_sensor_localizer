FROM alpine:3.8

ENV MQTT_ADDRESS localhost
ENV MQTT_PORT 1883
ENV MQTT_USER ""
ENV MQTT_PASSWORD ""
ENV RTL_433_MQTT_TOPIC rtl_433

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> etc/apk/repositories && \
    apk add --no-cache --virtual build-deps gcc libc-dev make cmake autoconf libusb-dev librtlsdr-dev libtool git && \
    apk add --no-cache libusb librtlsdr tzdata && \
    mkdir /tmp/src && \
    cd /tmp/src && \
    git clone https://github.com/merbanan/rtl_433.git && \
    cd rtl_433 && \
    mkdir build && \
    cd build && \
    cmake .. && \
    make && \
    make install && \
    rm -r /tmp/src && \
    apk del build-deps

RUN mkdir /config
VOLUME /config

COPY ./low_energy_sensor_localizer.conf /config
COPY ./low_energy_sensor_localizer /usr/local/bin/
COPY launcher.sh launcher.sh

CMD sh launcher.sh
