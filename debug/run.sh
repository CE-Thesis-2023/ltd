#!/bin/bash

# Install dependencies
brew install hivemq/mqtt-cli/mqtt-cli

# Starts the server
cd ..
docker compose up -d

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "REGISTER CAMERA TO DEVICE"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_AddCamera","info":{"cameraId":"32845204","name":"Most Expensive One","ip":"192.168.8.55","port":0,"username":"admin","password":"bkcamera2023"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'


echo "=============================="
echo "GET DEVICE INFO"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetDeviceInfo","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET DEVICE STREAM CHANNELS"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetStreamChannels","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "START FFMPEG STREAM"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_StartFfmpegStream","info":{"cameraId":"32845204","channelId":"1"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET STREAM LIST"
echo "=============================="

curl -XGET 'http://localhost:8080/api/debug/streams'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET STREAM STATUS"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_GetStreamStatus","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET STREAMS (CLOUD SERVER)"
echo "=============================="

curl -XGET -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' \
    'http://103.165.142.44:7956/v1/vhosts/default/apps/camera/streams'

sleep '5s'
docker compose logs --since '5s'
echo

echo "=============================="
echo "END CAMERA STREAM"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t commands/$deviceId \
    -m '{"commandType":"Command_EndFfmpegStream","info":{"cameraId":"32845204"}}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "ROTATE CAMERA"
echo "=============================="

deviceId='ltdtestdevice'

mqtt pub -t ptzctrl/$deviceId \
    -m '{"cameraId":"32845204","pan":60,"tilt":0,"stopAfterSeconds":10}' \
    -h '103.165.142.44' \
    -p '9093'

sleep '5s'
docker compose logs --since '5s'

sleep '8s'

echo "=============================="
echo "END DEBUGGING"
echo "=============================="
docker compose down