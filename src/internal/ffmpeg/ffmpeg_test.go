package custff

import "testing"

func Test_FfmpegCommand(t *testing.T) {
	ffmpegCommand := NewFFmpegCommand()
	ffmpegCommand.
		WithSourceUrl("./security_camera_01.mp4").
		WithGlobalArguments(
			map[string]string{
				"hide_banner": "",
				"loglevel":    "info",
				"threads":     "2",
			},
		).
		WithInputArguments(map[string]string{
			"avoid_negative_ts":           "make_zero",
			"fflags":                      "+genpts+discardcorrupt",
			"rtsp_transport":              "tcp",
			"use_wallclock_as_timestamps": "1",
			"timeout":                     "5000000",
		}).
		WithDestinationUrl("srt://103.165.142.15:8890?streamid=publish:test_ffmpeg").
		WithOutputArguments(map[string]string{
			"f":        "mpegts",
			"c:v":      "libx264",
			"preset:v": "faster",
			"tune:v":   "zerolatency",
		}).
		WithScale(25, 1280, 720).
		WithHardwareAccelerationType("")

	res, err := ffmpegCommand.String()
	if err != nil {
		t.Error(err)
	}
	t.Log(res)
}
