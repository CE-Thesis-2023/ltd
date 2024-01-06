#!/bin/bash

curl -XPOST -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' -H "Content-type: application/json" -d '{
  "id": "ltdtest-000",
  "stream": {
    "name": "ltdtest",
    "variantNames": []
  },
  "protocol": "srt",
  "url": "srt://103.165.142.44:7958/camera/ltdtest?mode=caller&latency=120000&timeout=500000",
  "streamKey": ""
}' 'http://127.0.0.1:7956/v1/vhosts/default/apps/ltd:startPush'