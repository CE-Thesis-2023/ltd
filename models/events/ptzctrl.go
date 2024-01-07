package events

type PtzCtrlRequest struct {
	Pan              int  `json:"pan"`
	Tilt             int  `json:"tilt"`
	StopAfterSeconds *int `json:"stopAfterSeconds,omitempty"`
}
