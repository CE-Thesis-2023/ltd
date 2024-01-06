#!/bin/bash

curl -XGET \
    -H 'Referer: http://192.168.8.55/doc/page/config.asp' \
    -H 'Host: 192.168.8.55' \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
    'http://192.168.8.55/ISAPI/Image/channels'

curl -XGET \
    -H 'Referer: http://192.168.8.55/doc/page/config.asp' \
    -H 'Host: 192.168.8.55' \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
    'http://192.168.8.55/ISAPI/Event/capabilities'

curl -XGET \
    -H 'Referer: http://192.168.8.55/doc/page/config.asp' \
    -H 'Host: 192.168.8.55' \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
    'http://192.168.8.55/ISAPI/PTZCtrl/channels'

curl -XGET \
    -H 'Referer: http://192.168.8.55/doc/page/config.asp' \
    -H 'Host: 192.168.8.55' \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
    'http://192.168.8.55/ISAPI/PTZCtrl/channels/1/capabilities'

curl -XGET \
-H 'Referer: http://192.168.8.55/doc/page/config.asp' \
-H 'Host: 192.168.8.55' \
--cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
'http://192.168.8.55/ISAPI/PTZCtrl/channels/1/maxelevation/capabilities'

curl -XPUT \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
     -d '<?xml version="1.0" encoding="UTF-8"?>
        <PTZData><pan>60</pan><tilt>0</tilt></PTZData>' \
        'http://192.168.8.55/ISAPI/PTZCtrl/channels/1/continuous'

sleep '0.2'

curl -XPUT \
    --cookie 'WebSession_a0d89116a2=0682dd5f28959c0c784540337f9fcb01425c351e4fd4eb1bbf82b219333155ad' \
     -d '<?xml version="1.0" encoding="UTF-8"?>
        <PTZData><pan>0</pan><tilt>0</tilt></PTZData>' \
        'http://192.168.8.55/ISAPI/PTZCtrl/channels/1/continuous'