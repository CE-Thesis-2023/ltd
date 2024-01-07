#!/bin/bash

jsonRequest='{"commandType":"Command_AddCamera","info":{"cameraId":"32845204","name":"Most Expensive One","ip":"192.168.8.55","port":0,"username":"admin","password":"bkcamera2023"}}'
cameraId='32845204'

mqtt pub -t commands/$cameraId \
    -m "$jsonRequest" \
    -h 103.165.142.44 \
    -p 9093