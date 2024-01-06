package ome

import fastshot "github.com/opus-domini/fast-shot"

type OmeStatisticsApiInterface interface {
}

type omeStatisticsClient struct {
	restClient fastshot.ClientHttpMethods
}


