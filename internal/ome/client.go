package ome

import (
	"fmt"
	"labs/local-transcoder/internal/configs"
	"time"

	fastshot "github.com/opus-domini/fast-shot"
)

type OmeClientInterface interface {
	Push() OmePushApiInterface
	Statistics() OmeStatisticsApiInterface
	Streams() OmeStreamsApiInterface
}

type OmeClient struct {
	configs    *configs.OvenMediaEngineConfigs
	restClient fastshot.ClientHttpMethods
}

func NewOmeClient(configs *configs.OvenMediaEngineConfigs) OmeClientInterface {
	baseUrl := fmt.Sprintf("%s:%d", configs.Host, configs.Port)
	restClient := fastshot.NewClient(baseUrl).
		Auth().BasicAuth(configs.Username, configs.Password).
		Config().SetTimeout(time.Second * 5).
		Build()
	return &OmeClient{
		configs:    configs,
		restClient: restClient,
	}
}

func (c *OmeClient) Push() OmePushApiInterface {
	return &omePushApiClient{
		restClient:   c.restClient,
		localConfigs: c.configs,
	}
}

func (c *OmeClient) Statistics() OmeStatisticsApiInterface {
	return &omeStatisticsClient{
		restClient: c.restClient,
	}
}

func (c *OmeClient) Streams() OmeStreamsApiInterface {
	return &omeStreamsApiClient{
		localConfigs: c.configs,
		restClient:   c.restClient,
	}
}
