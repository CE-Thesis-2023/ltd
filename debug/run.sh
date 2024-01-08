#!/bin/bash

# Starts the server
cd ..
docker compose up -d

# Begin logs
docker compose logs -f &

sleep '5s'

# Simulate stream start
## Get device info
deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetDeviceInfo","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

## Add camera
deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_AddCamera","info":{"cameraId":"32845204","name":"Most Expensive One","ip":"192.168.8.55","port":0,"username":"admin","password":"bkcamera2023"}}' \
    -h '103.165.142.44' \
    -p '9093'

## Debug stream channels
deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetStreamChannels","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

## Start stream
deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_StartFfmpegStream","info":{"cameraId":"32845204","channelId":"1"}}' \
    -h '103.165.142.44' \
    -p '9093'

## Get stream list
sleep '5s'
curl -XGET 'http://localhost:8080/api/debug/streams'

## Get stream status
deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetStreamStatus","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'