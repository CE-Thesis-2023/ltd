package custrtsp

import (
	"time"

	rtsp "github.com/bluenviron/gortsplib/v4"
)

func New() *rtsp.Client {
	return &rtsp.Client{
		ReadTimeout: time.Second * 3,
	}
}
