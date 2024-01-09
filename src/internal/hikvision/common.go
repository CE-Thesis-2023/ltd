package hikvision

import (
	"encoding/xml"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	custhttp "github.com/CE-Thesis-2023/ltd/src/internal/http"

	fastshot "github.com/opus-domini/fast-shot"
)

type Status struct {
	ID            string `xml:"id,omitempty"`
	StatusCode    int    `xml:"statusCode"`
	StatusString  string `xml:"statusString"`
	SubStatusCode string `xml:"subStatusCode"`
}

type AdditionalError struct {
	StatusList []Status `xml:"Status"`
}

type ResponseStatus struct {
	XMLName       xml.Name         `xml:"ResponseStatus"`
	Version       string           `xml:"version,attr"`
	RequestURL    string           `xml:"requestURL"`
	StatusCode    int              `xml:"statusCode"`
	StatusString  string           `xml:"statusString"`
	ID            int              `xml:"id,omitempty"`
	SubStatusCode string           `xml:"subStatusCode"`
	ErrorCode     int              `xml:"errorCode,omitempty"`
	ErrorMsg      string           `xml:"errorMsg,omitempty"`
	AdditionalErr *AdditionalError `xml:"AdditionalErr,omitempty"`
}

func handleError(resp *fastshot.Response) error {
	switch resp.StatusCode() {
	case 400:
		var parsedResp ResponseStatus
		if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatInvalidArgument(parsedResp.StatusString)
	case 401:
		return custerror.ErrorPermissionDenied
	case 404:
		return custerror.ErrorNotFound
	case 409:
		var parsedResp ResponseStatus
		if err := custhttp.XMLResponse(resp, &parsedResp); err != nil {
			return err
		}
		return custerror.FormatAlreadyExists(parsedResp.StatusString)
	case 500:
		return custerror.ErrorInternal
	}

	return nil
}
