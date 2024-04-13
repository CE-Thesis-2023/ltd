package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"
	"github.com/CE-Thesis-2023/ltd/src/internal/logger"

	"go.uber.org/zap"
)

type PtzApiClientInterface interface {
	Channels(ctx context.Context) (*PTZCtrlChannelsResponse, error)
	Capabilities(ctx context.Context, channelId string) (*PTZChannelCapabilities, error)
	RawContinuous(ctx context.Context, req *PtzCtrlRawContinousRequest) error
	ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error
	Status(ctx context.Context, req *PtzCtrlStatusRequest) (*PTZStatus, error)
	Relative(ctx context.Context, req *PTZCtrlRelativeRequest) error
}

type ptzApiClient struct {
	httpClient *http.Client
	username   string
	password   string
	ip         string
}

func (c *ptzApiClient) getBaseUrl() string {
	return c.ip + "/PTZCtrl/channels"
}

func (c *ptzApiClient) getUrlWithChannel(id string) string {
	return fmt.Sprintf("%s/%s", c.getBaseUrl(), id)
}

type PTZCtrlChannelsResponse struct {
}

func (c *ptzApiClient) Channels(ctx context.Context) (*PTZCtrlChannelsResponse, error) {
	p, _ := url.Parse(c.getBaseUrl())

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if err := handleError(resp); err != nil {
		return nil, err
	}

	var parsedResp PTZCtrlChannelsResponse
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PTZChannelCapabilities struct {
	XMLName                        xml.Name           `xml:"PTZChanelCap" json:"-"`
	Version                        string             `xml:"version,attr" json:"-" `
	XMLNamespace                   string             `xml:"xmlns,attr" json:"-"`
	AbsolutePanTiltPositionSpace   *PanTiltPosition   `xml:"AbsolutePanTiltPositionSpace,omitempty" json:"absolutePanTiltPositionSpace,omitempty"`
	AbsoluteZoomPositionSpace      *ZoomPosition      `xml:"AbsoluteZoomPositionSpace,omitempty" json:"absoluteZoomPositionSpace,omitempty"`
	RelativePanTiltSpace           *PanTiltSpace      `xml:"RelativePanTiltSpace,omitempty" json:"relativePanTiltSpace,omitempty"`
	RelativeZoomSpace              *ZoomSpace         `xml:"RelativeZoomSpace,omitempty" json:"relativeZoomSpace,omitempty"`
	ContinuousPanTiltSpace         *PanTiltSpace      `xml:"ContinuousPanTiltSpace,omitempty" json:"continuousPanTiltSpace,omitempty"`
	ContinuousZoomSpace            *ZoomSpace         `xml:"ContinuousZoomSpace,omitempty" json:"continuousZoomSpace,omitempty"`
	MomentaryPanTiltSpace          *PanTiltSpace      `xml:"MomentaryPanTiltSpace,omitempty" json:"momentaryPanTiltSpace,omitempty" `
	MomentaryZoomSpace             *ZoomSpace         `xml:"MomentaryZoomSpace,omitempty" json:"momentaryZoomSpace,omitempty"`
	HomePositionSupport            bool               `xml:"homePostionSupport" json:"homePositionSupport"`
	MaxPresetNum                   int                `xml:"maxPresetNum" json:"maxPresetNum"`
	MaxPatrolNum                   int                `xml:"maxPatrolNum" json:"maxPatrolNum"`
	MaxPatternNum                  int                `xml:"maxPatternNum" json:"maxPatternNum"`
	MaxLimitesNum                  int                `xml:"maxLimitesNum" json:"maxLimitesNum"`
	MaxTimeTaskNum                 int                `xml:"maxTimeTaskNum" json:"maxTimeTaskNum"`
	SerialNumber                   SerialNumber       `xml:"serialNumber" json:"serialNumber"`
	ControlProtocol                string             `xml:"controlProtocol" json:"controlProtocol"`
	ControlAddress                 string             `xml:"controlAddress" json:"controlAddress"`
	PTZRs485Para                   *PTZRs485Para      `xml:"PTZRs485Para,omitempty" json:"PTZRs485Para,omitempty"`
	PresetNameCap                  *PresetNameCap     `xml:"PresetNameCap,omitempty" json:"PresetNameCap,omitempty"`
	IsSupportPosition3D            bool               `xml:"isSupportPosition3D" json:"isSupportPosition3D"`
	IsSupportManualTrack           bool               `xml:"isSupportManualTrack" json:"isSupportManualTrack"`
	ManualControlSpeed             string             `xml:"manualControlSpeed" json:"manualControlSpeed"`
	IsSupportPtzlimiteds           bool               `xml:"isSupportPtzlimiteds" json:"isSupportPtzlimiteds"`
	ParkAction                     *ParkAction        `xml:"ParkAction,omitempty" json:"ParkAction,omitempty"`
	TimeTaskList                   *TimeTaskList      `xml:"TimeTaskList,omitempty" json:"TimeTaskList,omitempty"`
	Thermometry                    *Thermometry       `xml:"Thermometry,omitempty" json:"Thermometry,omitempty"`
	IsSupportPtzEagleFocusing      bool               `xml:"isSupportPtzEagleFocusing" json:"isSupportPtzEagleFocusing"`
	TrackingRatio                  *TrackingRatio     `xml:"TrackingRatio,omitempty" json:"TrackingRatio,omitempty"`
	TrackInitPosition              *TrackInitPosition `xml:"TrackInitPosition,omitempty" json:"TrackInitPosition,omitempty"`
	IsSupportAbsoluteEx            bool               `xml:"isSupportAbsoluteEx" json:"isSupportAbsoluteEx"`
	IsSupportCruise                bool               `xml:"isSupportCruise" json:"isSupportCruise"`
	IsSupportAreaScan              bool               `xml:"isSupportAreaScan" json:"isSupportAreaScan"`
	IsSupportFaceSnap3D            bool               `xml:"isSupportFaceSnap3D" json:"isSupportFaceSnap3D"`
	IsSupportOnepushSynchronizeFOV bool               `xml:"isSupportOnepushSynchronizeFOV" json:"isSupportOnepushSynchronizeFOV"`
	IsSupportLensCorrection        bool               `xml:"isSupportLensCorrection" json:"isSupportLensCorrection"`
	IsSupportPTZTrackStatus        bool               `xml:"isSupportPTZTrackStatus" json:"isSupportPTZTrackStatus"`
	PqrsZoom                       *PqrsZoom          `xml:"pqrsZoom,omitempty" json:"pqrsZoom,omitempty"`
	MnstFocus                      *MnstFocus         `xml:"mnstFocus,omitempty" json:"mnstFocus,omitempty"`
	IsSupportPTZSave               bool               `xml:"isSupportPTZSave" json:"isSupportPTZSave"`
	IsSupportPTZSaveGet            bool               `xml:"isSupportPTZSaveGet" json:"isSupportPTZSaveGet"`
	IsSupportAutoGotoCfg           bool               `xml:"isSupportAutoGotoCfg" json:"isSupportAutoGotoCfg"`
	LockTime                       int                `xml:"lockTime" json:"lockTime"`
}

type PanTiltPosition struct {
	XRange *XRange `xml:"XRange,omitempty" json:"XRange,omitempty"`
	YRange *YRange `xml:"YRange,omitempty" json:"YRange,omitempty"`
}

type ZoomPosition struct {
	ZRange *ZRange `xml:"ZRange,omitempty" json:"ZRange,omitempty"`
}

type PanTiltSpace struct {
	XRange *XRange `xml:"XRange,omitempty" json:"XRange,omitempty"`
	YRange *YRange `xml:"YRange,omitempty" json:"YRange,omitempty"`
}

type ZoomSpace struct {
	ZRange *ZRange `xml:"ZRange,omitempty" json:"ZRange,omitempty"`
}

type SerialNumber struct {
	Min int `xml:"min,attr" json:"min"`
	Max int `xml:"max,attr" json:"max"`
}

type PTZRs485Para struct {
	BaudRate   int    `xml:"baudRate" json:"baudRate"`
	DataBits   int    `xml:"dataBits" json:"dataBits"`
	ParityType string `xml:"parityType" json:"parityType" `
	StopBits   string `xml:"stopBits" json:"stopBits"`
	FlowCtrl   string `xml:"flowCtrl" json:"flowCtrl"`
}

type PresetNameCap struct {
	PresetNameSupport bool `xml:"presetNameSupport" json:"presetNameSupport"`
}

type ParkAction struct {
	AutoParkAction bool `xml:"autoParkAction" json:"autoParkAction"`
	GotoPresetNum  int  `xml:"gotoPresetNum" json:"gotoPresetNum"`
}

type TimeTaskList struct {
	TaskNum int         `xml:"taskNum" json:"taskNum"`
	Tasks   []*TimeTask `xml:"Task" json:"tasks"`
}

type TimeTask struct {
	TaskID           int           `xml:"taskID" json:"taskID"`
	TaskName         string        `xml:"taskName" json:"taskName"`
	TaskType         string        `xml:"taskType" json:"taskType"`
	TaskEnable       bool          `xml:"taskEnable" json:"taskEnable"`
	StartDateTime    string        `xml:"startDateTime" json:"startDateTime" `
	EndDateTime      string        `xml:"endDateTime" json:"endDateTime"`
	RepeatType       string        `xml:"repeatType" json:"repeatType"`
	IntervalType     string        `xml:"intervalType" json:"intervalType" `
	IntervalDuration int           `xml:"intervalDuration" json:"intervalDuration"`
	TriggerType      string        `xml:"triggerType" json:"triggerType"`
	TriggerValue     string        `xml:"triggerValue" json:"triggerValue"`
	Actions          []*TaskAction `xml:"Actions" json:"actions"`
}

type TaskAction struct {
	ActionType   string `xml:"actionType" json:"actionType"`
	ActionValue  string `xml:"actionValue" json:"actionValue"`
	ActionParams string `xml:"actionParams" json:"actionParams"`
}

type Thermometry struct {
	ThermometryCap *ThermometryCap `xml:"ThermometryCap,omitempty" json:"thermometryCap,omitempty"`
}

type ThermometryCap struct {
	EmissivityRange          *EmissivityRange          `xml:"EmissivityRange,omitempty" json:"emissivityRange,omitempty"`
	RelativeHumidityRange    *RelativeHumidityRange    `xml:"RelativeHumidityRange,omitempty" json:"relativeHumidityRange,omitempty" `
	AtmosphericPressureRange *AtmosphericPressureRange `xml:"AtmosphericPressureRange,omitempty" json:"atmosphericPressureRange,omitempty"`
	TemperatureRange         *TemperatureRange         `xml:"TemperatureRange,omitempty" json:"temperatureRange,omitempty" `
}

type EmissivityRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type RelativeHumidityRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type AtmosphericPressureRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type TemperatureRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type TrackingRatio struct {
	Max int `xml:"max,attr" `
}

type TrackInitPosition struct {
	PositionX float64 `xml:"positionX" json:"positionX" `
	PositionY float64 `xml:"positionY" json:"positionY"`
	ZoomValue int     `xml:"zoomValue" json:"zoomValue"`
}

type PqrsZoom struct {
	ZoomSpeedList []*ZoomSpeed `xml:"ZoomSpeedList" json:"zoomSpeedList"`
}

type ZoomSpeed struct {
	ZoomSpeedLevel string `xml:"zoomSpeedLevel" json:"zoomSpeedLevel"`
	ZoomSpeedValue int    `xml:"zoomSpeedValue" json:"zoomSpeedValue"`
}

type MnstFocus struct {
	MinstFocusSpeedList []*FocusSpeed `xml:"MinstFocusSpeedList"`
}

type FocusSpeed struct {
	FocusSpeedLevel string `xml:"focusSpeedLevel" json:"focusSpeedLevel"`
	FocusSpeedValue int    `xml:"focusSpeedValue" json:"focusSpeedValue"`
}

type XRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type YRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

type ZRange struct {
	Min float64 `xml:"min,attr"`
	Max float64 `xml:"max,attr"`
}

func (c *ptzApiClient) Capabilities(ctx context.Context, channelId string) (*PTZChannelCapabilities, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/capabilities", c.getUrlWithChannel(channelId)))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	if err := handleError(resp); err != nil {
		return nil, err
	}

	var parsedResp PTZChannelCapabilities
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PtzCtrlRawContinousRequest struct {
	ChannelId string
	Options   *PtzCtrlContinousOptions
}

type PtzCtrlContinousOptions struct {
	XMLName xml.Name `xml:"PTZData"`
	Pan     int      `xml:"pan"`  // + is right, - is left
	Tilt    int      `xml:"tilt"` // + is up, - is down
}

func (c *ptzApiClient) RawContinuous(ctx context.Context, req *PtzCtrlRawContinousRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId)))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req.Options),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}

	if err := handleError(resp); err != nil {
		return err
	}

	return nil
}

type PtzCtrlContinousWithResetRequest struct {
	ChannelId  string
	Options    *PtzCtrlContinousOptions
	ResetAfter time.Duration
}

var ptzCtrlResetRequestBody PtzCtrlContinousOptions = PtzCtrlContinousOptions{
	Pan:  0,
	Tilt: 0,
}

func (c *ptzApiClient) ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId)))

	resetRequest, err := custhttp.NewHttpRequest(
		context.Background(),
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(ptzCtrlResetRequestBody),
	)
	if err != nil {
		return err
	}

	continousRequest, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req.Options),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(continousRequest)
	if err != nil {
		return err
	}

	if err := handleError(resp); err != nil {
		return err
	}

	go func() {
		<-time.After(req.ResetAfter)
		logger.SDebug("PTZ Control continuous request start")

		resp, err := c.httpClient.Do(resetRequest)
		if err != nil {
			logger.SError("failed to send PTZ Control continuous reset request",
				zap.Error(err))
			return
		}

		if err := handleError(resp); err != nil {
			logger.SError("failed to send PTZ Control continuous reset request",
				zap.Error(err))
			return
		}

		logger.SDebug("PTZ Control continuous request reset completed",
			zap.String("channelId", req.ChannelId),
			zap.Duration("after", req.ResetAfter))
	}()

	return nil
}

type PtzCtrlStatusRequest struct {
	ChannelId string
}

type PTZStatus struct {
	XMLName      xml.Name     `xml:"PTZStatus"`
	Version      string       `xml:"version,attr"`
	AbsoluteHigh AbsoluteHigh `xml:"AbsoluteHigh"`
}

type AbsoluteHigh struct {
	Elevation    int `xml:"elevation"`
	Azimuth      int `xml:"azimuth"`
	AbsoluteZoom int `xml:"absoluteZoom"`
}

func (c *ptzApiClient) Status(ctx context.Context, req *PtzCtrlStatusRequest) (*PTZStatus, error) {
	p, _ := url.Parse(fmt.Sprintf("%s/status", c.getUrlWithChannel(req.ChannelId)))

	statusRequest, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodGet,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req),
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(statusRequest)
	if err != nil {
		return nil, err
	}

	if err := handleError(resp); err != nil {
		return nil, err
	}

	var parsedResp PTZStatus
	if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type PTZCtrlRelativeRequest struct {
	XMLName  xml.Name `xml:"PTZData"`
	Relative Relative `xml:"Relative"`
}
type Relative struct {
	PositionX    float32 `xml:"positionX"`
	PositionY    float32 `xml:"positionY"`
	RelativeZoom float32 `xml:"relativeZoom"`
}

func (c *ptzApiClient) Relative(ctx context.Context, req *PTZCtrlRelativeRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/relative", c.getUrlWithChannel("1")))
	logger.SDebug("PTZ relative request",
		zap.String("url", p.String()))

	request, err := custhttp.NewHttpRequest(
		ctx,
		p,
		http.MethodPut,
		custhttp.WithBasicAuth(c.username, c.password),
		custhttp.WithContentType("application/xml"),
		custhttp.WithXMLBody(req),
	)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		logger.SDebug("PTZ relative request failed", zap.Error(err))
		return err
	}

	if err := handleError(resp); err != nil {
		logger.SDebug("PTZ relative request failed", zap.Error(err))
		return err
	}

	return nil
}
