package hikvision

import (
	"context"
	"encoding/xml"
	"fmt"
	custhttp "labs/local-transcoder/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
	"github.com/panjf2000/ants/v2"
)

type StreamsApiInterface interface {
	Channels(ctx context.Context, req *StreamChannelsRequest) (*StreamingChannelList, error)
	Status(ctx context.Context, req *StreamingStatusRequest) (*StreamingStatusResponse, error)
}

type streamApiClient struct {
	restClient fastshot.ClientHttpMethods
	pool       *ants.Pool
}

func (c *streamApiClient) getBaseUrl() string {
	return "/Stream"
}

type StreamChannelsRequest struct {
	ChannelId string `json:"channelId"`
}

type StreamingChannelList struct {
	XMLName          xml.Name           `xml:"StreamingChannelList"`
	StreamingChannel []StreamingChannel `xml:"StreamingChannel,omitempty"`
}

func (c *streamApiClient) Channels(ctx context.Context, req *StreamChannelsRequest) (*StreamingChannelList, error) {
	url := fmt.Sprintf("%s/channels", c.getBaseUrl())
	if req.ChannelId != "" {
		url = fmt.Sprintf("%s/%s", url, req.ChannelId)
	}

	resp, err := c.restClient.GET(url).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp StreamingChannelList
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

func (c *streamApiClient) Status(ctx context.Context, req *StreamingStatusRequest) (*StreamingStatusResponse, error) {
	url := fmt.Sprintf("%s/status", c.getBaseUrl())

	resp, err := c.restClient.GET(url).
		Context().Set(ctx).
		Send()
	if err != nil {
		return nil, err
	}

	if err := handleError(&resp); err != nil {
		return nil, err
	}

	var parsedResp StreamingStatusResponse
	if err := custhttp.XMLResponse(&resp, &parsedResp); err != nil {
		return nil, err
	}

	return &parsedResp, nil
}

type StreamingStatusRequest struct {
}

type StreamingStatusResponse struct {
	XMLName                    xml.Name                   `xml:"StreamingStatus"`
	TotalStreamingSessions     int                        `xml:"totalStreamingSessions"`
	StreamingSessionStatusList StreamingSessionStatusList `xml:"StreamingSessionStatusList"`
}

type StreamingSessionStatusList struct {
	StreamingSessionStatus []StreamingSessionStatus `xml:"StreamingSessionStatus"`
}

type StreamingSessionStatus struct {
	XMLName       xml.Name      `xml:"StreamingSessionStatus"`
	ClientAddress ClientAddress `xml:"clientAddress"`
}

type ClientAddress struct {
	IpAddress   string `xml:"ipAddress"`
	Ipv6Address string `xml:"ipv6Address"`
}

type StreamingChannel struct {
	XMLName            xml.Name       `xml:"StreamingChannel"`
	Version            string         `xml:"version,attr"`
	XMLNS              string         `xml:"xmlns,attr"`
	ID                 string         `xml:"id"`
	ChannelName        string         `xml:"channelName"`
	Enabled            bool           `xml:"enabled"`
	Transport          *Transport     `xml:"Transport,omitempty"`
	Unicast            *Unicast       `xml:"Unicast,omitempty"`
	Multicast          *Multicast     `xml:"Multicast,omitempty"`
	Security           *Security      `xml:"Security,omitempty"`
	SRTPMulticast      *SRTPMulticast `xml:"SRTPMulticast,omitempty"`
	Video              *Video         `xml:"Video,omitempty"`
	Audio              *Audio         `xml:"Audio,omitempty"`
	EnableCABAC        *bool          `xml:"enableCABAC,omitempty"`
	SubStreamRecStatus *bool          `xml:"subStreamRecStatus,omitempty"`
	CustomStreamEnable *bool          `xml:"customStreamEnable,omitempty"`
}

type Transport struct {
	XMLName                  xml.Name            `xml:"Transport"`
	MaxPacketSize            *int                `xml:"maxPacketSize,omitempty"`
	AudioPacketLength        *int                `xml:"audioPacketLength,omitempty"`
	AudioInboundPacketLength *int                `xml:"audioInboundPacketLength,omitempty"`
	AudioInboundPortNo       *int                `xml:"audioInboundPortNo,omitempty"`
	VideoSourcePortNo        *int                `xml:"videoSourcePortNo,omitempty"`
	AudioSourcePortNo        *int                `xml:"audioSourcePortNo,omitempty"`
	ControlProtocolList      ControlProtocolList `xml:"ControlProtocolList"`
}

type ControlProtocolList struct {
	ControlProtocols []ControlProtocol `xml:"ControlProtocol"`
}

type ControlProtocol struct {
	StreamingTransport string `xml:"streamingTransport"`
}

type Unicast struct {
	XMLName          xml.Name `xml:"Unicast"`
	Enabled          bool     `xml:"enabled"`
	InterfaceID      *string  `xml:"interfaceID,omitempty"`
	RTPTransportType string   `xml:"rtpTransportType,omitempty"`
}

type Multicast struct {
	XMLName                xml.Name `xml:"Multicast"`
	Enabled                bool     `xml:"enabled"`
	UserTriggerThreshold   *int     `xml:"userTriggerThreshold,omitempty"`
	DestIPAddress          *string  `xml:"destIPAddress,omitempty"`
	VideoDestPortNo        *int     `xml:"videoDestPortNo,omitempty"`
	AudioDestPortNo        *int     `xml:"audioDestPortNo,omitempty"`
	DestIPv6Address        *string  `xml:"destIPv6Address,omitempty"`
	TTL                    *int     `xml:"ttl,omitempty"`
	ActiveMulticastEnabled *bool    `xml:"activeMulticastEnabled,omitempty"`
	PackagingFormat        *string  `xml:"packagingFormat,omitempty"`
	FecInfo                *FecInfo `xml:"FecInfo,omitempty"`
}

type FecInfo struct {
	XMLName       xml.Name `xml:"FecInfo"`
	FecRatio      int      `xml:"fecRatio"`
	FecDestPortNo *int     `xml:"fecDestPortNo,omitempty"`
}

type Security struct {
	XMLName         xml.Name `xml:"Security"`
	Enabled         bool     `xml:"enabled"`
	CertificateType string   `xml:"certificateType"`
}

type SRTPMulticast struct {
	XMLName             xml.Name `xml:"SRTPMulticast"`
	SRTPVideoDestPortNo *int     `xml:"SRTPVideoDestPortNo,omitempty"`
	SRTPAudioDestPortNo *int     `xml:"SRTPAudioDestPortNo,omitempty"`
}

type Video struct {
	XMLName                 xml.Name    `xml:"Video"`
	Enabled                 bool        `xml:"enabled"`
	VideoInputChannelID     string      `xml:"videoInputChannelID"`
	VideoCodecType          string      `xml:"videoCodecType"`
	VideoResolutionWidth    int         `xml:"videoResolutionWidth"`
	VideoResolutionHeight   int         `xml:"videoResolutionHeight"`
	VideoQualityControlType *string     `xml:"videoQualityControlType,omitempty"`
	ConstantBitRate         *int        `xml:"constantBitRate,omitempty"`
	VBRUpperCap             *int        `xml:"vbrUpperCap,omitempty"`
	VBRLowerCap             *int        `xml:"vbrLowerCap,omitempty"`
	MaxFrameRate            int         `xml:"maxFrameRate"`
	KeyFrameInterval        *int        `xml:"keyFrameInterval,omitempty"`
	RotationDegree          *int        `xml:"rotationDegree,omitempty"`
	MirrorEnabled           *bool       `xml:"mirrorEnabled,omitempty"`
	SnapShotImageType       *string     `xml:"snapShotImageType,omitempty"`
	Mpeg4Profile            *string     `xml:"Mpeg4Profile,omitempty"`
	H264Profile             *string     `xml:"H264Profile,omitempty"`
	SVACProfile             *string     `xml:"SVACProfile,omitempty"`
	GovLength               *int        `xml:"GovLength,omitempty"`
	SVC                     *SVC        `xml:"SVC"`
	Smoothing               *int        `xml:"smoothing,omitempty"`
	SmartCodec              *SmartCodec `xml:"SmartCodec,omitempty"`
	VBRAverageCap           *int        `xml:"vbrAverageCap,omitempty"`
}

type SmartCodec struct {
	Enabled bool   `xml:"enabled,omitempty"`
	SVCMode string `xml:"SVCMode,omitempty"`
}

type SVC struct {
	Smoothing int `xml:"smoothing,omitempty"`
}

type Audio struct {
	Enabled                    bool          `xml:"enabled"`
	AudioInputChannelID        string        `xml:"audioInputChannelID"`
	AudioCompressionType       string        `xml:"audioCompressionType,omitempty"`
	AudoInboundCompressionType *string       `xml:"audioInboundCompressionType,omitempty"`
	AudioBitRate               *int          `xml:"audioBitRate,omitempty"`
	AudioSamplingRate          *int          `xml:"audioSamplingRate,omitempty"`
	AudioResolution            *bool         `xml:"audoResolution,omitempty"`
	VoiceChanger               *VoiceChanger `xml:"VoiceChanger,omitempty"`
}

type VoiceChanger struct {
	Enabled bool `xml:"enabled"`
	Level   int  `xml:"level"`
}
