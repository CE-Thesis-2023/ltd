#!/bin/bash

# Stream desktop to OME
./ffmpeg -f x11grab -follow_mouse centered -r 25 -s cif -i :0.0 -c:v libx264 \
       -f mpegts 'srt://103.165.142.44:7958?streamid=srt%3A%2F%2F103.165.142.44%3A7958%2Fcamera%2F32845204'

# Restream RTMP to SRT to OME
ffmpeg -i 'rtsp://rtspstream:b82358d74a99f22c0941d57b0d5857d7@zephyr.rtsp.stream/movie' -c copy -f mpegts 'srt://103.165.142.44:7958?streamid=srt%3A%2F%2F103.165.142.44%3A7958%2Fcamera%2F32845204'
