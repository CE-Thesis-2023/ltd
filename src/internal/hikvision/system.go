package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
)

type SystemApiInterface interface {
	Capabilities(ctx context.Context) (*SystemCapabilitiesResponse, error)
	DeviceInfo(ctx context.Context) (*SystemDeviceInfoResponse, error)
	Hardware(ctx context.Context) (*SystemHardwareResponse, error)
	Status(ctx context.Context) (*SystemStatus, error)
}

type systemApiClient struct {
	restClient fastshot.ClientHttpMethods
}

func (c *systemApiClient) getBaseUrl() string {
	return "/System"
}

type SystemCapabilitiesResponse struct {
}

func (c *systemApiClient) Capabilities(ctx context.Context) (*SystemCapabilitiesResponse, error) {
	p := fmt.Sprintf("%s/capabilities", c.getBaseUrl())

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemCapabilitiesResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

func (c *systemApiClient) DeviceInfo(ctx context.Context) (*SystemDeviceInfoResponse, error) {
	p := fmt.Sprintf("%s/deviceinfo", c.getBaseUrl())

	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemDeviceInfoResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type SystemHardwareResponse struct {
}

func (c *systemApiClient) Hardware(ctx context.Context) (*SystemHardwareResponse, error) {
	p := fmt.Sprintf("%s/Hardware", c.getBaseUrl())
	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemHardwareResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

func (c *systemApiClient) Status(ctx context.Context) (*SystemStatus, error) {
	p := fmt.Sprintf("%s/status", c.getBaseUrl())
	resp, err := c.restClient.GET(p).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp SystemStatus
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type SystemDeviceInfoResponse struct {
	XMLName              xml.Name `xml:"DeviceInfo"`
	Version              string   `xml:"version,attr"`
	XMLNS                string   `xml:"xmlns,attr"`
	DeviceName           string   `xml:"deviceName"`
	DeviceID             string   `xml:"deviceID"`
	DeviceDescription    string   `xml:"deviceDescription,omitempty"`
	DeviceLocation       string   `xml:"deviceLocation,omitempty"`
	DeviceStatus         DeviceStatus
	SystemContact        string  `xml:"systemContact,omitempty"`
	Model                string  `xml:"model"`
	SerialNumber         string  `xml:"serialNumber"`
	MacAddress           string  `xml:"macAddress"`
	FirmwareVersion      string  `xml:"firmwareVersion"`
	FirmwareReleasedDate string  `xml:"firmwareReleasedDate,omitempty"`
	BootVersion          string  `xml:"bootVersion,omitempty"`
	BootReleasedDate     string  `xml:"bootReleasedDate,omitempty"`
	HardwareVersion      string  `xml:"hardwareVersion,omitempty"`
	EncoderVersion       string  `xml:"encoderVersion,omitempty"`
	EncoderReleasedDate  string  `xml:"encoderReleasedDate,omitempty"`
	DecoderVersion       string  `xml:"decoderVersion,omitempty"`
	DecoderReleasedDate  string  `xml:"decoderReleasedDate,omitempty"`
	SoftwareVersion      string  `xml:"softwareVersion,omitempty"`
	Capacity             int     `xml:"capacity,omitempty"`
	UsedCapacity         int     `xml:"usedCapacity,omitempty"`
	DeviceType           string  `xml:"deviceType"`
	TelecontrolID        int     `xml:"telecontrolID,omitempty"`
	SupportBeep          bool    `xml:"supportBeep,omitempty"`
	FirmwareVersionInfo  string  `xml:"firmwareVersionInfo,omitempty"`
	ActualFloorNum       int     `xml:"actualFloorNum"`
	SubChannelEnabled    bool    `xml:"subChannelEnabled,omitempty"`
	ThrChannelEnabled    bool    `xml:"thrChannelEnabled,omitempty"`
	RadarVersion         string  `xml:"radarVersion,omitempty"`
	LocalZoneNum         int     `xml:"localZoneNum,omitempty"`
	AlarmOutNum          int     `xml:"alarmOutNum,omitempty"`
	DistanceResolution   float64 `xml:"distanceResolution,omitempty"`
	AngleResolution      float64 `xml:"angleResolution,omitempty"`
	SpeedResolution      float64 `xml:"speedResolution,omitempty"`
	DetectDistance       float64 `xml:"detectDistance,omitempty"`
	LanguageType         string  `xml:"languageType,omitempty"`
	RelayNum             int     `xml:"relayNum,omitempty"`
	ElectroLockNum       int     `xml:"electroLockNum,omitempty"`
	RS485Num             int     `xml:"RS485Num,omitempty"`
	PowerOnMode          string  `xml:"powerOnMode,omitempty"`
}

type DeviceStatus struct {
	Status               string `xml:"deviceStatus,omitempty"`
	DetailAbnormalStatus DetailAbnormalStatus
}

type DetailAbnormalStatus struct {
	HardDiskFull         bool `xml:"hardDiskFull,omitempty"`
	HardDiskError        bool `xml:"hardDiskError,omitempty"`
	EthernetBroken       bool `xml:"ethernetBroken,omitempty"`
	IPAddrConflict       bool `xml:"ipaddrConflict,omitempty"`
	IllegalAccess        bool `xml:"illegalAccess,omitempty"`
	RecordError          bool `xml:"recordError,omitempty"`
	RAIDLogicDiskError   bool `xml:"raidLogicDiskError,omitempty"`
	SpareWorkDeviceError bool `xml:"spareWorkDeviceError,omitempty"`
}

type SystemStatus struct {
	XMLName               string                 `xml:"DeviceStatus"`
	Version               string                 `xml:"version,attr"`
	XMLNamespace          string                 `xml:"xmlns,attr"`
	CurrentDeviceTime     *time.Time             `xml:"currentDeviceTime,omitempty"`
	DeviceUpTime          *int                   `xml:"deviceUpTime,omitempty"`
	TemperatureList       *TemperatureList       `xml:"TemperatureList,omitempty"`
	FanList               *FanList               `xml:"FanList,omitempty"`
	PressureList          *PressureList          `xml:"PressureList,omitempty"`
	TamperList            *TamperList            `xml:"TamperList,omitempty"`
	CPUList               *CPUList               `xml:"CPUList,omitempty"`
	MemoryList            *MemoryList            `xml:"MemoryList,omitempty"`
	OpenFileHandles       *int                   `xml:"openFileHandles,omitempty"`
	CameraList            *CameraList            `xml:"CameraList,omitempty"`
	DomeInfoList          *DomeInfoList          `xml:"DomeInfoList,omitempty"`
	DeviceStatus          string                 `xml:"deviceStatus"`
	DialSignalStrength    *int                   `xml:"dialSignalStrength,omitempty"`
	USBStatusList         []USBStatus            `xml:"USBStatusList>USBStatus"`
	WifiStatusList        []WifiStatus           `xml:"WifiStatusList>WifiStatus"`
	AlertStreamServerList *AlertStreamServerList `xml:"AlertStreamServerList,omitempty"`
}

type TemperatureList struct {
	Temperature []Temperature `xml:"Temperature,omitempty"`
}

type Temperature struct {
	TempSensorDescription string  `xml:"tempSensorDescription"`
	Temperature           float64 `xml:"temperature"`
}

type FanList struct {
	Fan []Fan `xml:"Fan,omitempty"`
}

type Fan struct {
	FanDescription string `xml:"fanDescription"`
	Speed          int    `xml:"speed"`
}

type PressureList struct {
	Pressure []Pressure `xml:"Pressure,omitempty"`
}

type Pressure struct {
	PressureSensorDescription string `xml:"pressureSensorDescription"`
	Pressure                  int    `xml:"pressure"`
}

type TamperList struct {
	Tamper []Tamper `xml:"Tamper,omitempty"`
}

type Tamper struct {
	TamperSensorDescription string `xml:"tamperSensorDescription"`
	Tamper                  bool   `xml:"tamper"`
}

type CPUList struct {
	CPU []CPU `xml:"CPU,omitempty"`
}

type CPU struct {
	CPUDescription string `xml:"cpuDescription"`
	CPUUtilization int    `xml:"cpuUtilization"`
}

type MemoryList struct {
	Memory []Memory `xml:"Memory,omitempty"`
}

type Memory struct {
	MemoryDescription string  `xml:"memoryDescription"`
	MemoryUsage       float64 `xml:"memoryUsage"`
	MemoryAvailable   float64 `xml:"memoryAvailable"`
}

type CameraList struct {
	Camera []Camera `xml:"Camera,omitempty"`
}

type Camera struct {
	ZoomReverseTimes   int `xml:"zoomReverseTimes"`
	ZoomTotalSteps     int `xml:"zoomTotalSteps"`
	FocusReverseTimes  int `xml:"focusReverseTimes"`
	FocusTotalSteps    int `xml:"focusTotalSteps"`
	IrisShiftTimes     int `xml:"irisShiftTimes"`
	IrisTotalSteps     int `xml:"irisTotalSteps"`
	IcrShiftTimes      int `xml:"icrShiftTimes"`
	IcrTotalSteps      int `xml:"icrTotalSteps"`
	LensIntirTimes     int `xml:"lensIntirTimes"`
	CameraRunTotalTime int `xml:"cameraRunTotalTime"`
}

type DomeInfoList struct {
	DomeInfo []DomeInfo `xml:"DomeInfo,omitempty"`
}

type DomeInfo struct {
	DomeRunTotalTime            *int `xml:"domeRunTotalTime,omitempty"`
	RunTimeUnderNegativetwenty  *int `xml:"runTimeUnderNegativetwenty,omitempty"`
	RunTimeBetweenNtwentyPforty *int `xml:"runTimeBetweenNtwentyPforty,omitempty"`
	RuntimeOverPositiveforty    *int `xml:"runtimeOverPositiveforty,omitempty"`
	PanTotalRounds              *int `xml:"panTotalRounds,omitempty"`
	TiltTotalRounds             *int `xml:"tiltTotalRounds,omitempty"`
	HeatState                   *int `xml:"heatState,omitempty"`
	FanState                    *int `xml:"fanState,omitempty"`
}

type USBStatus struct {
	ID    int    `xml:"id"`
	State string `xml:"state,omitempty"`
}

type WifiStatus struct {
	ID    int    `xml:"id"`
	State string `xml:"state,omitempty"`
}

type AlertStreamServerList struct {
	AlertStreamServer []AlertStreamServer `xml:"AlertStreamServer,omitempty"`
}

type AlertStreamServer struct {
	ID           *int    `xml:"id,omitempty"`
	ProtocolType *string `xml:"protocolType,omitempty"`
	IPAddress    *string `xml:"ip"`
}
