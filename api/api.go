package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

const splitioAPIUri = "https://sdk.split.io/api"

// SplitioAPIBinding contains splitioAPIKey
type SplitioAPIBinding struct {
	splitioAPIKey string
	splitioAPIUri string
}

// NewSplitioAPIBinding returns a new SplitioAPIBinding
func NewSplitioAPIBinding(apiKey string, apiURL string) *SplitioAPIBinding {
	if apiURL == "" {
		apiURL = splitioAPIUri
	}
	return &SplitioAPIBinding{apiKey, apiURL}
}

// HTTPGet makes a GET request to the Split.io SDK API.
// path is the path of the HTTP request e.g "/splitChanges"
// since is an integer used as a query string, will be -1 on the first request
func (binding *SplitioAPIBinding) httpGet(path string, since int) (map[string]interface{}, error) {
	client := resty.New()
	resp, err := client.R().
		SetPathParams(map[string]string{
			"path": path,
		}).
		SetQueryParams(map[string]string{
			"since": string(since),
		}).
		SetHeaders(map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", binding.splitioAPIKey),
		}).
		Get(fmt.Sprintf("%s/{path}", binding.splitioAPIUri))

	if err != nil {
		err = fmt.Errorf("Http get request error: %s", err)
		return map[string]interface{}{}, err
	}

	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("Non-OK HTTP status: %s", resp.Status())
		return map[string]interface{}{}, err
	}

	var data map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(resp.Body())))
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if err != nil {
		err = fmt.Errorf("Decode error: %s", err)
		return map[string]interface{}{}, err
	}

	return data, nil
}

// GetSegmentChanges will get segment data
func (binding *SplitioAPIBinding) GetSegmentChanges() error {
	return fmt.Errorf("not implemented")
}

// GetSplitChanges will get split data
func (binding *SplitioAPIBinding) GetSplitChanges() error {
	return fmt.Errorf("not implemented")
}
