#!/bin/bash

BACKEND_HOST=http://103.165.142.44:7880

# Starts the server
cd ..
docker compose up -d

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET LIST TRANSCODERS"
echo "=============================="

deviceId='ltdtestdevice'

curl --insecure -XGET $BACKEND_HOST/api/devices
echo
sleep '5s'

echo "=============================="
echo "GET LIST CAMERAS"
echo "=============================="

deviceId='ltdtestdevice'

curl --insecure -XGET $BACKEND_HOST/api/cameras
echo
sleep '5s'

echo "=============================="
echo "REGISTER CAMERA TO DEVICE"
echo "=============================="

deviceId='ltdtestdevice'

cameraId=$(curl --insecure -XPOST \
    -d '{"name":"Expensive Camera","ip":"192.168.8.55","port":0,"username":"admin","password":"bkcamera2023","transcoderId":"ltdtestdevice"}' \
    $BACKEND_HOST/api/cameras | jq -r '.cameraId')

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "START FFMPEG STREAM"
echo "=============================="

deviceId='ltdtestdevice'

curl --insecure -XPUT \
    $BACKEND_HOST/api/cameras/$cameraId/streams?enable=true
echo

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET CAMERA LIST"
echo "=============================="

curl --insecure -XGET $BACKEND_HOST/api/cameras
echo

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET CAMERA STATUS"
echo "=============================="

deviceId='ltdtestdevice'

curl --insecure -XGET $BACKEND_HOST/api/cameras?id=$cameraId
echo

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "GET STREAMS (CLOUD SERVER)"
echo "=============================="

curl --insecure -XGET $BACKEND_HOST/api/cameras/$cameraId/streams
echo

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "ROTATE CAMERA"
echo "=============================="

deviceId='ltdtestdevice'

reqRotateCamera01='{"cameraId":"'
reqRotateCamera02='","pan":60,"tilt":30}'
reqRotateCamera="$reqRotateCamera01$cameraId$reqRotateCamera02"
echo $reqRotateCamera
curl --insecure -XPOST \
    -d $reqRotateCamera \
    $BACKEND_HOST/api/rc
echo

sleep '5s'
docker compose logs --since '5s'

sleep '8s'

echo "=============================="
echo "DELETE CAMERA"
echo "=============================="

deviceId='ltdtestdevice'

curl --insecure -XDELETE $BACKEND_HOST/api/cameras?id=$cameraId
echo

sleep '5s'
docker compose logs --since '5s'

echo "=============================="
echo "END DEBUGGING"
echo "=============================="
docker compose down