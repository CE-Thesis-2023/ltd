#!/bin/bash

jsonRequest='{"cameraId":"32845204","pan":0,"tilt":-60,"stopAfterSeconds":2}'
cameraId='32845204'

mqtt pub -t ptzctrl/$cameraId \
    -m "$jsonRequest" \
    -h 103.165.142.44 \
    -p 9093