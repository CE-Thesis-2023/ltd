package events

type PtzCtrlRequest struct {
	CameraId         string `json:"cameraId"`
	Pan              int    `json:"pan"`
	Tilt             int    `json:"tilt"`
	StopAfterSeconds *int   `json:"stopAfterSeconds,omitempty"`
}
