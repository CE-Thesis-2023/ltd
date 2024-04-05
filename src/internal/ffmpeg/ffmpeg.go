package custff

import (
	"fmt"
	"strings"
)

type FFmpegHardwareAccelerationType string

var (
	VA_API    FFmpegHardwareAccelerationType = "vaapi"
	QUICKSYNC FFmpegHardwareAccelerationType = "quicksync"
)

type ffmpegCommand struct {
	sourceUrl                string
	destinationUrl           string
	inputArguments           map[string]string
	outputArguments          map[string]string
	globalArguments          map[string]string
	hardwareAccelerationType FFmpegHardwareAccelerationType
	fps                      int
	width                    int
	height                   int
	binPath                  string
}

func NewFFmpegCommand() *ffmpegCommand {
	return &ffmpegCommand{}
}

func (c *ffmpegCommand) WithBinPath(path string) *ffmpegCommand {
	c.binPath = path
	return c
}

func (c *ffmpegCommand) WithInputArguments(args map[string]string) *ffmpegCommand {
	c.inputArguments = args
	return c
}

func (c *ffmpegCommand) WithOutputArguments(args map[string]string) *ffmpegCommand {
	c.outputArguments = args
	return c
}

func (c *ffmpegCommand) WithGlobalArguments(args map[string]string) *ffmpegCommand {
	c.globalArguments = args
	return c
}

func (c *ffmpegCommand) WithSourceUrl(url string) *ffmpegCommand {
	c.sourceUrl = url
	return c
}

func (c *ffmpegCommand) WithDestinationUrl(url string) *ffmpegCommand {
	c.destinationUrl = url
	return c
}

func (c *ffmpegCommand) WithHardwareAccelerationType(t FFmpegHardwareAccelerationType) *ffmpegCommand {
	c.hardwareAccelerationType = t
	return c
}

func (c *ffmpegCommand) WithScale(fps int, width int, height int) *ffmpegCommand {
	c.fps = fps
	c.width = width
	c.height = height
	return c
}

func (c *ffmpegCommand) String() (string, error) {
	cmd := "ffmpeg"
	if c.binPath != "" {
		cmd = c.binPath
	}
	if c.globalArguments != nil {
		cmd += " " + c.toArguments(c.globalArguments)
	}
	cmd += c.buildDecodeHardwareArguments()
	if c.inputArguments != nil {
		cmd += " " + c.toArguments(c.inputArguments)
	}
	if c.sourceUrl == "" {
		return "", fmt.Errorf("source URL is required")
	}
	cmd += " -i " + fmt.Sprintf("'%s'", c.sourceUrl)
	if c.fps > 0 && c.width > 0 && c.height > 0 {
		cmd += c.buildScaleHardwareArguments(c.fps, c.width, c.height)
	}
	if c.outputArguments != nil {
		cmd += " " + c.toArguments(c.outputArguments)
	}
	if c.destinationUrl == "" {
		return "", fmt.Errorf("destination URL is required")
	}
	cmd += " " + fmt.Sprintf("'%s'", c.destinationUrl)
	return cmd, nil
}

func (c *ffmpegCommand) buildDecodeHardwareArguments() string {
	switch c.hardwareAccelerationType {
	case VA_API:
		return " -hwaccel vaapi -hwaccel_flags allow_profile_mismatch -hwaccel_device /dev/dri/renderD128 -hwaccel_output_format vaapi"
	case QUICKSYNC:
		return "-hwaccel qsv -qsv_device /dev/dri/renderD128 -hwaccel_output_format qsv -c:v h264_qsv"
	default:
		return ""
	}
}

func (c *ffmpegCommand) buildScaleHardwareArguments(fps int, width int, height int) string {
	switch c.hardwareAccelerationType {
	case VA_API:
		return fmt.Sprintf(" -r %d -vf fps=%d,scale_vaapi=w=%d:h=%d:format=nv12,hwdownload,format=nv12,format=yuv420p",
			fps, fps, width, height)
	case QUICKSYNC:
		return fmt.Sprintf(" -r %d -vf vpp_qsv=framerate=%d:w=%d:h=%d:format=nv12,hwdownload,format=nv12,format=yuv420p",
			fps, fps, width, height)
	default:
		return fmt.Sprintf(" -r %d -vf scale=%d:%d", fps, width, height)
	}
}

func (c *ffmpegCommand) toArguments(a map[string]string) string {
	var args []string
	for k, v := range a {
		args = append(args, fmt.Sprintf("-%s", k))
		if v != "" {
			args = append(args, v)
		}
	}
	return strings.Join(args, " ")
}
