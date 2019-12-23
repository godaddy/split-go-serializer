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
	allChanges, since, err := binding.getAllChanges(path)
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

// GetSegmentsForSplits return segment info and the count of splits using segment
func (binding *SplitioAPIBinding) GetSegmentsForSplits(splits []dtos.SplitDTO) ([]dtos.SegmentChangesDTO, int, error) {
	allSegmentNames := map[string]bool{}
	segments := []dtos.SegmentChangesDTO{}
	usingSegmentsCount := 0

	for _, split := range splits {
		segmentNames := getSegmentNamesInUse(split.Conditions)
		if len(segmentNames) > 0 {
			usingSegmentsCount++
		}
		for segmentName := range segmentNames {
			allSegmentNames[segmentName] = true
		}
	}

	for segmentName := range allSegmentNames {
		segment, err := binding.getSegment(segmentName)
		if err != nil {
			return segments, 0, err
		}
		segments = append(segments, segment)

	}

	return segments, usingSegmentsCount, nil
}

// httpGet makes a GET request to the Split.io SDK API.
// path is the path of the HTTP request, either "splitChanges" or "segmentChanges/segmentName"
// since is an integer used as a query string, will be -1 on the first request
func (binding *SplitioAPIBinding) httpGet(path string, since int64) (map[string]interface{}, error) {
	client := resty.New()
	resp, err := client.R().
		SetQueryParams(map[string]string{
			"since": strconv.FormatInt(since, 10),
		}).
		SetHeaders(map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", binding.splitioAPIKey),
		}).
		Get(fmt.Sprintf("%s/%s", binding.splitioAPIUri, path))

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
// path is the path of the HTTP request e.g "splitChanges", "segmentChanges/segmentName"
func (binding *SplitioAPIBinding) getAllChanges(path string) ([]map[string]interface{}, int64, error) {
	since := firstRequestSince
	requestCount := 0
	allChanges := []map[string]interface{}{}
	for requestCount < defaultMaxRequestNum {
		results, err := binding.httpGet(path, since)
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

func getSegmentNamesInUse(conditions []dtos.ConditionDTO) map[string]bool {
	segmentNames := map[string]bool{}
	for _, condition := range conditions {
		for _, matcher := range condition.MatcherGroup.Matchers {
			if matcher.MatcherType == "IN_SEGMENT" {
				segmentNames[matcher.UserDefinedSegment.SegmentName] = true
			}
		}
	}

	return segmentNames

}

// Get info for single segment
func (binding *SplitioAPIBinding) getSegment(segmentName string) (dtos.SegmentChangesDTO, error) {
	path := "segmentChanges"
	segment := dtos.SegmentChangesDTO{}
	allChanges, since, err := binding.getAllChanges(fmt.Sprintf("%s/%s", path, segmentName))
	if err != nil {
		return segment, err
	}

	addedMap := map[string]bool{}
	for _, changes := range allChanges {
		var SegmentChanges dtos.SegmentChangesDTO
		err = mapstructure.Decode(changes, &SegmentChanges)
		if err != nil {
			err = fmt.Errorf("error when decode data to segment: %s", err)
			return segment, err
		}

		for _, add := range SegmentChanges.Added {
			addedMap[add] = true
		}

		for _, remove := range SegmentChanges.Removed {
			delete(addedMap, remove)
		}

	}

	adds := []string{}
	for add := range addedMap {
		adds = append(adds, add)
	}

	segment = dtos.SegmentChangesDTO{
		Name:  segmentName,
		Added: adds,
		Since: since,
		Till:  since,
	}

	return segment, nil

}
