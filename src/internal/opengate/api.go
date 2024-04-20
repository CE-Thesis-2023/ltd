package opengate

import (
	"context"
	"fmt"
	"io"
	"net/http"

	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
)

type OpenGateHTTPAPIClient struct {
	baseUrl    string
	httpClient *http.Client
}

func NewOpenGateHTTPAPIClient(baseUrl string) *OpenGateHTTPAPIClient {
	return &OpenGateHTTPAPIClient{
		baseUrl:    baseUrl,
		httpClient: http.DefaultClient,
	}
}

func (c *OpenGateHTTPAPIClient) EventsSnapshot(ctx context.Context, eventId string, height int, quality int) ([]byte, error) {
	uri := fmt.Sprintf("%s/api/events/%s/snapshot.jpg", c.baseUrl, eventId)
	if height > 0 {
		uri = fmt.Sprintf("%s?height=%d", uri, height)
	}
	if quality > 0 {
		uri = fmt.Sprintf("%s&quality=%d", uri, quality)
	}
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri,
		nil)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to create http request: %s", err)
	}
	resp, err := c.
		httpClient.
		Do(req)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to send http request: %s", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to read response body: %s", err)
	}
	return bodyBytes, nil
}

func (c *OpenGateHTTPAPIClient) EventsThumbnail(ctx context.Context, eventId string) ([]byte, error) {
	uri := fmt.Sprintf("%s/api/events/%s/thumbnail.jpg", c.baseUrl, eventId)
	req, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri,
		nil)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to create http request: %s", err)
	}
	resp, err := c.
		httpClient.
		Do(req)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to send http request: %s", err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, custerror.FormatInternalError("failed to read response body: %s", err)
	}
	return bodyBytes, nil
}
