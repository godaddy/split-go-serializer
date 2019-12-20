package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/stretchr/testify/assert"
)

const (
	mockSplitioAPIKey = "someKey"
	mockSplitioAPIURI = "https://mock.sdk.split.io/api"
)

func TestNewSplitioAPIBindingValid(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(mockSplitioAPIKey, mockSplitioAPIURI)

	// Validate that returned NewSplitioAPIBinding has the correct api key
	assert.EqualValues(t, result.splitioAPIKey, mockSplitioAPIKey)
	assert.EqualValues(t, result.splitioAPIUri, mockSplitioAPIURI)
}

func TestNewSplitioAPIBindingHasDefaultURI(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(mockSplitioAPIKey, "")

	// Validate that returned NewSplitioAPIBinding has the default uri
	assert.EqualValues(t, result.splitioAPIUri, "https://sdk.split.io/api")
}

func TestHttpGetReturnsSuccessfulResponse(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"data":"fake splitio json string"}`)
	}))
	defer testServer.Close()

	path := "mockPath"
	segment := "mockSegment"
	since := -1

	// Act
	expectedData := map[string]interface{}{
		"data": "fake splitio json string",
	}
	binding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := binding.httpGet(path, segment, since)

	// Validate that httpGet function returns correct data and empty error
	assert.Equal(t, result, expectedData)
	assert.Nil(t, err)
}

func TestHttpGetReturnsErrorOnNonOKResponse(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()

	path := "mockPath"
	segment := "mockSegment"
	since := -1

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := apiBinding.httpGet(path, segment, since)

	// Validate that httpGet function returns unsuccessful error
	assert.EqualError(t, err, "Non-OK HTTP status: 404 Not Found")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestHttpGetReturnsNewRequestError(t *testing.T) {
	// Arrange
	badURI := ":"
	path := "mockPath"
	segment := "mockSegment"
	since := -1

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, badURI)
	result, err := apiBinding.httpGet(path, segment, since)

	// Validate that httpGet function returns new request error
	assert.EqualError(t, err, "Http get request error: parse :/mockPath/mockSegment: missing protocol scheme")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestHttpGetReturnsDecodeError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid-response`)
	}))
	defer testServer.Close()

	path := "mockPath"
	segment := "mockSegment"
	since := -1

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := apiBinding.httpGet(path, segment, since)

	// Validate that httpGet function returns new request error
	assert.EqualError(t, err, "Decode error: invalid character 'i' looking for beginning of value")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestGetSegmentChangesReturnsError(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(mockSplitioAPIKey, mockSplitioAPIURI)
	err := result.GetSegmentChanges()

	// Validate that GetSegmentChanges function returns error
	assert.EqualError(t, err, "not implemented")
}

func TestGetSplitChangesReturnsError(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(mockSplitioAPIKey, mockSplitioAPIURI)
	err := result.GetSplitChanges()

	// Validate that GetSplitChanges function returns error
	assert.EqualError(t, err, "not implemented")
}

func TestGetSegmentNamesInUseValid(t *testing.T) {
	// Arrange
	mockConditions := []byte(`
	[{
          "conditionType": "foo",
          "matcherGroup": {
            "matchers": [
              {
                "matcherType": "WHITELIST"
              }
            ]
          }
        },
        {
          "conditionType": "bar",
          "matcherGroup": {
            "matchers": [
              {
                "matcherType": "IN_SEGMENT",
                "userDefinedSegmentMatcherData": {
                  "segmentName": "test-segment"
                }
              }
            ]
          }
        }
      ]`)
	conditions := []dtos.ConditionDTO{}
	json.Unmarshal(mockConditions, &conditions)

	// Act
	segmentNames := getSegmentNamesInUse(conditions)
	expectedNames := map[string]bool{
		"test-segment": true,
	}

	// Validate that returned segmentNames has the correct names
	assert.Equal(t, segmentNames, expectedNames)
	assert.Equal(t, segmentNames["test-segment"], true)
}
