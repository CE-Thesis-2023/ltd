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
}

func (c *ptzApiClient) getBaseUrl() string {
	return "/PTZCtrl/channels"
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
	XMLName                        xml.Name           `xml:"PTZChannelCap"`
	Version                        string             `xml:"version,attr"`
	XMLNamespace                   string             `xml:"xmlns,attr"`
	AbsolutePanTiltPositionSpace   *PanTiltPosition   `xml:"AbsolutePanTiltPositionSpace,omitempty"`
	AbsoluteZoomPositionSpace      *ZoomPosition      `xml:"AbsoluteZoomPositionSpace,omitempty"`
	RelativePanTiltSpace           *PanTiltSpace      `xml:"RelativePanTiltSpace,omitempty"`
	RelativeZoomSpace              *ZoomSpace         `xml:"RelativeZoomSpace,omitempty"`
	ContinuousPanTiltSpace         *PanTiltSpace      `xml:"ContinuousPanTiltSpace,omitempty"`
	ContinuousZoomSpace            *ZoomSpace         `xml:"ContinuousZoomSpace,omitempty"`
	MomentaryPanTiltSpace          *PanTiltSpace      `xml:"MomentaryPanTiltSpace,omitempty"`
	MomentaryZoomSpace             *ZoomSpace         `xml:"MomentaryZoomSpace,omitempty"`
	HomePositionSupport            bool               `xml:"homePostionSupport"`
	MaxPresetNum                   int                `xml:"maxPresetNum"`
	MaxPatrolNum                   int                `xml:"maxPatrolNum"`
	MaxPatternNum                  int                `xml:"maxPatternNum"`
	MaxLimitesNum                  int                `xml:"maxLimitesNum"`
	MaxTimeTaskNum                 int                `xml:"maxTimeTaskNum"`
	SerialNumber                   SerialNumber       `xml:"serialNumber"`
	ControlProtocol                string             `xml:"controlProtocol,omitempty"`
	ControlAddress                 string             `xml:"controlAddress,omitempty"`
	PTZRs485Para                   *PTZRs485Para      `xml:"PTZRs485Para,omitempty"`
	PresetNameCap                  *PresetNameCap     `xml:"PresetNameCap,omitempty"`
	IsSupportPosition3D            bool               `xml:"isSupportPosition3D,omitempty"`
	IsSupportManualTrack           bool               `xml:"isSupportManualTrack,omitempty"`
	ManualControlSpeed             string             `xml:"manualControlSpeed,omitempty"`
	IsSupportPtzlimiteds           bool               `xml:"isSpportPtzlimiteds,omitempty"`
	ParkAction                     *ParkAction        `xml:"ParkAction,omitempty"`
	TimeTaskList                   *TimeTaskList      `xml:"TimeTaskList,omitempty"`
	Thermometry                    *Thermometry       `xml:"Thermometry,omitempty"`
	IsSupportPtzEagleFocusing      bool               `xml:"isSpportPtzEagleFocusing,omitempty"`
	TrackingRatio                  *TrackingRatio     `xml:"TrackingRatio,omitempty"`
	TrackInitPosition              *TrackInitPosition `xml:"TrackInitPosition,omitempty"`
	IsSupportAbsoluteEx            bool               `xml:"isSupportAbsoluteEx,omitempty"`
	IsSupportCruise                bool               `xml:"isSupportCruise,omitempty"`
	IsSupportAreaScan              bool               `xml:"isSupportAreaScan,omitempty"`
	IsSupportFaceSnap3D            bool               `xml:"isSupportFaceSnap3D,omitempty"`
	IsSupportOnepushSynchronizeFOV bool               `xml:"isSupportOnepushSynchronizeFOV,omitempty"`
	IsSupportLensCorrection        bool               `xml:"isSupportLensCorrection,omitempty"`
	IsSupportPTZTrackStatus        bool               `xml:"isSupportPTZTrackStatus,omitempty"`
	PqrsZoom                       *PqrsZoom          `xml:"pqrsZoom,omitempty"`
	MnstFocus                      *MnstFocus         `xml:"mnstFocus,omitempty"`
	IsSupportPTZSave               bool               `xml:"isSupportPTZSave,omitempty"`
	IsSupportPTZSaveGet            bool               `xml:"isSupportPTZSaveGet,omitempty"`
	IsSupportAutoGotoCfg           bool               `xml:"isSupportAutoGotoCfg,omitempty"`
	LockTime                       int                `xml:"lockTime,omitempty"`
}

type PanTiltPosition struct {
	XRange *XRange `xml:"XRange,omitempty"`
	YRange *YRange `xml:"YRange,omitempty"`
}

type ZoomPosition struct {
	ZRange *ZRange `xml:"ZRange,omitempty"`
}

type PanTiltSpace struct {
	XRange *XRange `xml:"XRange,omitempty"`
	YRange *YRange `xml:"YRange,omitempty"`
}

type ZoomSpace struct {
	ZRange *ZRange `xml:"ZRange,omitempty"`
}

type SerialNumber struct {
	Min int `xml:"min,attr"`
	Max int `xml:"max,attr"`
}

type PTZRs485Para struct {
	BaudRate   int    `xml:"baudRate"`
	DataBits   int    `xml:"dataBits"`
	ParityType string `xml:"parityType"`
	StopBits   string `xml:"stopBits"`
	FlowCtrl   string `xml:"flowCtrl"`
}

type PresetNameCap struct {
	PresetNameSupport bool `xml:"presetNameSupport"`
}

type ParkAction struct {
	AutoParkAction bool `xml:"autoParkAction"`
	GotoPresetNum  int  `xml:"gotoPresetNum"`
}

type TimeTaskList struct {
	TaskNum int         `xml:"taskNum"`
	Tasks   []*TimeTask `xml:"Task"`
}

type TimeTask struct {
	TaskID           int           `xml:"taskID"`
	TaskName         string        `xml:"taskName"`
	TaskType         string        `xml:"taskType"`
	TaskEnable       bool          `xml:"taskEnable"`
	StartDateTime    string        `xml:"startDateTime"`
	EndDateTime      string        `xml:"endDateTime"`
	RepeatType       string        `xml:"repeatType"`
	IntervalType     string        `xml:"intervalType"`
	IntervalDuration int           `xml:"intervalDuration"`
	TriggerType      string        `xml:"triggerType"`
	TriggerValue     string        `xml:"triggerValue"`
	Actions          []*TaskAction `xml:"Actions>TaskAction"`
}

type TaskAction struct {
	ActionType   string `xml:"actionType"`
	ActionValue  string `xml:"actionValue"`
	ActionParams string `xml:"actionParams"`
}

type Thermometry struct {
	ThermometryCap *ThermometryCap `xml:"ThermometryCap,omitempty"`
}

type ThermometryCap struct {
	EmissivityRange          *EmissivityRange          `xml:"EmissivityRange,omitempty"`
	RelativeHumidityRange    *RelativeHumidityRange    `xml:"RelativeHumidityRange,omitempty"`
	AtmosphericPressureRange *AtmosphericPressureRange `xml:"AtmosphericPressureRange,omitempty"`
	TemperatureRange         *TemperatureRange         `xml:"TemperatureRange,omitempty"`
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
	Max int `xml:"max,attr"`
}

type TrackInitPosition struct {
	PositionX float64 `xml:"positionX"`
	PositionY float64 `xml:"positionY"`
	ZoomValue int     `xml:"zoomValue"`
}

type PqrsZoom struct {
	ZoomSpeedList []*ZoomSpeed `xml:"ZoomSpeedList>ZoomSpeed"`
}

type ZoomSpeed struct {
	ZoomSpeedLevel string `xml:"zoomSpeedLevel"`
	ZoomSpeedValue int    `xml:"zoomSpeedValue"`
}

type MnstFocus struct {
	MinstFocusSpeedList []*FocusSpeed `xml:"MinstFocusSpeedList>FocusSpeed"`
}

type FocusSpeed struct {
	FocusSpeedLevel string `xml:"focusSpeedLevel"`
	FocusSpeedValue int    `xml:"focusSpeedValue"`
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

var ptzCtrlResetRequestBody string = "<PTZData><pan>0</pan><tilt>0</tilt></PTZData>"

func (c *ptzApiClient) ContinousWithReset(ctx context.Context, req *PtzCtrlContinousWithResetRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/continuous", c.getUrlWithChannel(req.ChannelId)))

	resetRequest, err := custhttp.NewHttpRequest(
		ctx,
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
	PositionX    int `xml:"positionX"`
	PositionY    int `xml:"positionY"`
	RelativeZoom int `xml:"relativeZoom"`
}

func (c *ptzApiClient) Relative(ctx context.Context, req *PTZCtrlRelativeRequest) error {
	p, _ := url.Parse(fmt.Sprintf("%s/relative", c.getBaseUrl()))

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
		return err
	}

	if err := handleError(resp); err != nil {
		return err
	}

	return nil
}
