package events

type CommandType string

const (
	Command_GetDeviceInfo     CommandType = "Command_GetDeviceInfo"
	Command_StartStream       CommandType = "Command_StartStream"
	Command_EndStream         CommandType = "Command_EndStream"
	Command_AddCamera         CommandType = "Command_AddCamera"
	Command_StartFfmpegStream CommandType = "Command_StartFfmpegStream"
	Command_EndFfmpegStream   CommandType = "Command_EndFfmpegStream"
	Command_GetStreamChannels CommandType = "Command_GetStreamChannels"
	Command_GetStreamStatus   CommandType = "Command_GetStreamStatus"
)

type CommandRequest struct {
	CommandType CommandType            `json:"commandType"`
	Info        map[string]interface{} `json:"info"`
}

type CommandRetrieveDeviceInfo struct {
	CameraId         string `json:"cameraId"`
}

type CommandRetrieveStreamChannels struct {
	CameraId string `json:"channelId"`
}

type CommandStartStreamInfo struct {
	CameraId  string `json:"cameraId"`
	ChannelId string `json:"channelId"`
}

type CommandAddCameraInfo struct {
	CameraId string `json:"cameraId"`
	Name     string `json:"name"`
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CommandEndStreamInfo struct {
	CameraId string `json:"cameraId"`
}

type CommandGetStreamStatusRequest struct {
	CameraId string `json:"cameraId"`
}
