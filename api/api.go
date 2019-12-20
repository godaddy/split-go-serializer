package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/splitio/go-client/splitio/service/dtos"
)

const (
	splitioAPIUri        = "https://sdk.split.io/api"
	firstRequestSince    = int64(-1)
	defaultMaxRequestNum = 100
)

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

// GetSplits gets the split data
func (binding *SplitioAPIBinding) GetSplits() ([]dtos.SplitDTO, int64, error) {
	path := "splitChanges"
	splitsMap := map[string]dtos.SplitDTO{}
	splits := []dtos.SplitDTO{}
	allChanges, since, err := binding.getAllChanges(path, "")
	if err != nil {
		return nil, 0, err
	}

	for _, changes := range allChanges {
		var result dtos.SplitChangesDTO
		err = mapstructure.Decode(changes, &result)
		if err != nil {
			err = fmt.Errorf("error when decode data to split: %s", err)
			return nil, 0, err
		}

		for _, split := range result.Splits {
			if split.Status == "ARCHIVED" {
				delete(splitsMap, split.Name)
			} else {
				splitsMap[split.Name] = split
			}
		}
	}

	for _, split := range splitsMap {
		splits = append(splits, split)
	}
	return splits, since, nil
}

// GetSegmentChanges will get segment data
func (binding *SplitioAPIBinding) GetSegmentChanges() error {
	return fmt.Errorf("not implemented")
}

// httpGet makes a GET request to the Split.io SDK API.
// path is the path of the HTTP request, either "splitChanges" or "segmentChanges"
// segment is the segment name when path is "segmentChange", otherwise should be empty
// since is an integer used as a query string, will be -1 on the first request
func (binding *SplitioAPIBinding) httpGet(path string, segment string, since int64) (map[string]interface{}, error) {
	client := resty.New()
	resp, err := client.R().
		SetPathParams(map[string]string{
			"path":    path,
			"segment": segment,
		}).
		SetQueryParams(map[string]string{
			"since": strconv.FormatInt(since, 10),
		}).
		SetHeaders(map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", binding.splitioAPIKey),
		}).
		Get(fmt.Sprintf("%s/{path}/{segment}", binding.splitioAPIUri))

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

// getAllChanges polls the Split.io API until since and till are the same
// path is the path of the HTTP request e.g "splitChanges", "segmentChanges"
// segment is the segment name, will be empty when the end point is splitChanges
func (binding *SplitioAPIBinding) getAllChanges(path string, segment string) ([]map[string]interface{}, int64, error) {
	since := firstRequestSince
	requestCount := 0
	allChanges := []map[string]interface{}{}
	for requestCount < defaultMaxRequestNum {
		results, err := binding.httpGet(path, segment, since)
		if err != nil {
			return nil, 0, err
		}
		till, err := results["till"].(json.Number).Int64()
		if err != nil {
			return nil, 0, err
		}

		if since == till {
			break
		}
		allChanges = append(allChanges, results)
		since = till
		requestCount++
	}

	return allChanges, since, nil
}
