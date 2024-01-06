#!/bin/bash

# curl -XGET \
#     --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
#     'http://192.168.8.55/ISAPI/Smart/capabilities'

curl -XGET \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
    'http://192.168.8.55/ISAPI/Streaming/status'