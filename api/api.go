package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	// initiate a new request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", binding.splitioAPIUri, path), nil)
	if err != nil {
		err := fmt.Errorf("Errored when initiate http new request. %s", err)
		return map[string]interface{}{}, err
	}

	// add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", binding.splitioAPIKey))

	// add query string
	q := req.URL.Query()
	q.Add("since", string(since))
	req.URL.RawQuery = q.Encode()

	// send http request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err := fmt.Errorf("Errored when sending request to the server %s", err)
		return map[string]interface{}{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Non-OK HTTP status: %s", resp.Status)
		return map[string]interface{}{}, err
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	var data map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(respBody)))
	decoder.UseNumber()
	decoder.Decode(&data)

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
