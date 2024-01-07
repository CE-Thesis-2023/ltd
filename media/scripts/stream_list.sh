#!/bin/bash

curl -XGET -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' \
    'http://127.0.0.1:7956/v1/vhosts/default/apps/camera/streams'

# TO GET ALL STREAM

## FOR LTD

curl -XGET -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' \
    'http://127.0.0.1:7956/v1/vhosts/default/apps/ltd/streams'

## FOR CLOUD 

curl -XGET -H 'Authorization: Basic dGhlc2lzOnExamsyM2kxOQ==' \
    'http://127.0.0.1:7956/v1/vhosts/default/apps/cloud/streams'