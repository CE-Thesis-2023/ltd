package ms

import (
	"fmt"
	"net/url"
	"time"
)

type Protocol string

const (
	ProtocolWebRTC    Protocol = "webrtc"
	ProtocolRTMP      Protocol = "rtmp"
	ProtocolSRT       Protocol = "srt"
	ProtocolLLHLS     Protocol = "llhls"
	ProtocolThumbnail Protocol = "thumbnail"
)

type Status string

const (
	StatusOpening Status = "opening"
	StatusClosing Status = "closing"
)

type Direction string

const (
	DirectionIncoming = "incoming"
	DirectionOutgoing = "outgoing"
)

type AdmissionWebhookRequest struct {
	Client  AdmissionClientInfo    `json:"client"`
	Request AdmissionClientRequest `json:"request"`
}

type AdmissionClientInfo struct {
	Address   string `json:"address"`
	Port      int    `json:"port"`
	UserAgent string `json:"user_agent"`
}

type AdmissionClientRequest struct {
	Direction Direction  `json:"direction"`
	Protocol  Protocol   `json:"protocol"`
	Status    Status     `json:"status"`
	URL       string     `json:"url"`
	NewURL    string     `json:"new_url"`
	Time      *time.Time `json:"time"`
}

func (r *AdmissionClientRequest) ParseUrl() (*url.URL, error) {
	if r.URL == "" {
		return nil, fmt.Errorf("URL is empty")
	}

	parsedURL, err := url.Parse(r.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return parsedURL, nil
}

func (r *AdmissionClientRequest) ParseNewUrl() (*url.URL, error) {
	if r.URL == "" {
		return nil, fmt.Errorf("NewURL is empty")
	}

	parsedURL, err := url.Parse(r.NewURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NewURL: %w", err)
	}

	return parsedURL, nil
}

type AdmissionWebhookResponse struct {
	Allowed  *bool   `json:"allowed,omitempty"`
	NewURL   *string `json:"new_url,omitempty"`
	Lifetime *int    `json:"lifetime,omitempty"`
	Reason   *string `json:"reason,omitempty"`
}
