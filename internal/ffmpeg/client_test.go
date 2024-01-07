package cffmpeg_test

import (
	"testing"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func TestMain(t *testing.T) {
	ffmpeg.Input("rtsp://localhost:8554/smptebars").
		Output("srt://103.165.142.44:7958?streamid=srt%3A%2F%2F103.165.142.44%3A7958%2Fcamera%2F32845204", ffmpeg.KwArgs{
			"c:v":       "libx264",
			"c:a":       "aac",
			"f":         "mpegts",
			"preset":    "veryfast",
			"tune":      "zerolatency",
			"profile:v": "baseline",
			"s":         "1280x720",
			"filter:v":  "fps=30",
		}).ErrorToStdOut().Run()
}
