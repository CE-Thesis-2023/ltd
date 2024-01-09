package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	custhttp "github.com/CE-Thesis-2023/ltd/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
)

type SystemApiInterface interface {
	Capabilities(ctx context.Context) (*SystemCapabilitiesResponse, error)
	DeviceInfo(ctx context.Context) (*SystemDeviceInfoResponse, error)
	Hardware(ctx context.Context) (*SystemHardwareResponse, error)
}

type systemApiClient struct {
	restClient fastshot.ClientHttpMethods
	pool       *ants.Pool
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
