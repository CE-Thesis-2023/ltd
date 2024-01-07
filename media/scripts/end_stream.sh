#!/bin/bash

jsonRequest='{"commandType":"Command_EndFfmpegStream","info":{"cameraId":"32845204"}}'
cameraId='32845204'

mqtt pub -t commands/$cameraId \
    -m "$jsonRequest" \
    -h 103.165.142.44 \
    -p 9093