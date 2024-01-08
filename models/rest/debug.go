package rest

type DebugListStreamsResponse struct {
	Streams []StreamInfo `json:"streams"`
}

type StreamInfo struct {
	CameraId       string `json:"cameraId"`
	SourceUrl      string `json:"sourceUrl"`
	DestinationUrl string `json:"destinationUrl"`
}
