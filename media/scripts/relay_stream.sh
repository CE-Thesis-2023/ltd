#!/bin/bash

curl -XPOST -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' -H "Content-type: application/json" -d '{
  "id": "ltdtest-001",
  "stream": {
    "name": "ltdtest",
    "variantNames": []
  },
  "protocol": "srt",
  "url": "srt://103.165.142.44:7958?streamid=srt%3A%2F%2F103.165.142.44%3A7958%2Fcamera%2Fltdtest&mode=caller",
  "streamKey": ""
}' 'http://127.0.0.1:7956/v1/vhosts/default/apps/ltd:startPush'