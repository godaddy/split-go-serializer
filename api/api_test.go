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
	mockPath          = "mockPath"
	mockSince         = int64(-1)
	mockConditions    = `
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
                  "segmentName": "mock-segment"
                }
              }
            ]
          }
        }
    ]`
)

type mockHandler struct {
	count int
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI // e.g URI: "/splitCahges?since=-1"
	if path[:6] == "/split" {
		if h.count == 0 {
			fmt.Fprintln(w, `{"splits": [{"name":"mock-split-1", "killed": false},
										 {"name":"mock-split-2"}],
							  "since": -1, "till":10}`)
		} else if h.count == 1 {
			fmt.Fprintln(w, `{"splits": [{"name":"mock-split-1", "killed": true},
										 {"name":"mock-split-2", "status":"ARCHIVED"},
										 {"name":"mock-split-3"}, {"name":"mock-split-4"}],
							  "since": 10, "till":20}`)
		} else if h.count == 2 {
			fmt.Fprintln(w, `{"splits": [], "since":20, "till":20}`)
		}
	} else if path[:8] == "/segment" {
		if h.count == 0 {
			fmt.Fprintln(w, `{"name": "mock-segment", "added": ["mock1","mock2","mock3","mock4"],
			                  "removed": [], "since":-1, "till":5}`)
		} else if h.count == 1 {
			fmt.Fprintln(w, `{"name": "mock-segment", "added": ["mock5"], "removed": ["mock2"],
			                  "since":5, "till":10}`)
		} else if h.count == 2 {
			fmt.Fprintln(w, `{"name": "", "added": [], "removed": [],
			                  "since":10, "till":10}`)
		}
	}

	h.count++
}

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

	// Act
	expectedData := map[string]interface{}{
		"data": "fake splitio json string",
	}
	binding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := binding.httpGet(mockPath, mockSince)

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

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := apiBinding.httpGet(mockPath, mockSince)

	// Validate that httpGet function returns unsuccessful error
	assert.EqualError(t, err, "Non-OK HTTP status: 404 Not Found")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestHttpGetReturnsNewRequestError(t *testing.T) {
	// Arrange
	badURI := ":"

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, badURI)
	result, err := apiBinding.httpGet(mockPath, mockSince)

	// Validate that httpGet function returns new request error
	assert.EqualError(t, err, "Http get request error: parse :/mockPath: missing protocol scheme")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestHttpGetReturnsDecodeError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid-response`)
	}))
	defer testServer.Close()

	// Act
	apiBinding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)
	result, err := apiBinding.httpGet(mockPath, mockSince)

	// Validate that httpGet function returns new request error
	assert.EqualError(t, err, "Decode error: invalid character 'i' looking for beginning of value")
	assert.Equal(t, result, map[string]interface{}{})
}

func TestGetAllChangesValid(t *testing.T) {
	// Arrange
	handler := &mockHandler{}
	testServer := httptest.NewServer(handler)
	defer testServer.Close()
	binding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	changes, since, err := binding.getAllChanges("splitChanges")
	expectedSplits := []interface{}{
		map[string]interface{}{"name": "mock-split-1", "killed": false},
		map[string]interface{}{"name": "mock-split-2"},
	}

	// Valide that getAllChanges returns valid value
	assert.Nil(t, err)
	assert.Equal(t, since, int64(20))
	assert.Equal(t, changes[0]["splits"], expectedSplits)
	assert.Equal(t, len(changes), 2)
}

func TestGetAllChangesReturnsHTTPError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer testServer.Close()
	binding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	changes, since, err := binding.getAllChanges(mockPath)

	// Valide that getAllChanges return getHTTP error
	assert.EqualError(t, err, "Non-OK HTTP status: 404 Not Found")
	assert.Nil(t, changes)
	assert.Equal(t, since, int64(0))
}

func TestGetAllChangesReturnsIntConvertError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"till":3.15}`)
	}))
	defer testServer.Close()
	binding := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	changes, since, err := binding.getAllChanges(mockPath)

	// Valide that getAllChanges return parsing error
	assert.EqualError(t, err, "strconv.ParseInt: parsing \"3.15\": invalid syntax")
	assert.Nil(t, changes)
	assert.Equal(t, since, int64(0))
}

func TestGetSplitsValid(t *testing.T) {
	// Arrange
	handler := &mockHandler{}
	testServer := httptest.NewServer(handler)
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	splits, since, err := result.GetSplits()
	var splitOneKilled bool
	var splitTwoExist bool
	for _, split := range splits {
		if split.Name == "mock-split-1" {
			splitOneKilled = split.Killed
		}
		if split.Name == "mock-split-2" {
			splitTwoExist = true
		}
	}

	// Validate that GetSplits returns correct values
	assert.Equal(t, since, int64(20))
	assert.Equal(t, splitOneKilled, true)
	assert.Equal(t, splitTwoExist, false)
	assert.Equal(t, len(splits), 3)
	assert.Nil(t, err)
}

func TestGetSplitsReturnsGetAllChangesError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	splits, since, err := result.GetSplits()

	// Validate that GetSplits returns error from getAllChanges
	assert.Equal(t, since, int64(0))
	assert.Nil(t, splits)
	assert.EqualError(t, err, "Non-OK HTTP status: 401 Unauthorized")
}

func TestGetSplitsReturnsDecodeError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"since":"wrong-type", "till":10}`)
	}))
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	splits, since, err := result.GetSplits()

	// Validate that GetSplits returns decode error
	assert.Equal(t, since, int64(0))
	assert.Nil(t, splits)
	assert.EqualError(t, err, "error when decode data to split: 1 error(s) decoding:\n\n* 'Since' expected type 'int64', got unconvertible type 'string'")
}

func TestGetSegmentNamesInUseValid(t *testing.T) {
	// Arrange
	conditions := []dtos.ConditionDTO{}
	json.Unmarshal([]byte(mockConditions), &conditions)

	// Act
	segmentNames := getSegmentNamesInUse(conditions)
	expectedNames := map[string]bool{
		"mock-segment": true,
	}

	// Validate that returned segmentNames has the correct names
	assert.Equal(t, segmentNames, expectedNames)
	assert.Equal(t, segmentNames["mock-segment"], true)
}

func TestGetSegmentValid(t *testing.T) {
	// Arrange
	handler := &mockHandler{}
	testServer := httptest.NewServer(handler)
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	segment, err := result.getSegment("mock-segment")
	var valueFiveExists bool
	var valueTwoExists bool
	for _, added := range segment.Added {
		if added == "mock5" {
			valueFiveExists = true
		}
		if added == "mock2" {
			valueTwoExists = true
		}
	}

	// Validate that GetSegment function returns correct segment values
	assert.Equal(t, segment.Name, "mock-segment")
	assert.Equal(t, valueTwoExists, false)
	assert.Equal(t, valueFiveExists, true)
	assert.Nil(t, err)
}

func TestGetSegmentReturnsDecodeError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"since":"wrong-type", "till":10}`)
	}))
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	segment, err := result.getSegment("mock-segment-name")

	// Validate that GetSegment function returns decode error
	assert.EqualError(t, err, "error when decode data to segment: 1 error(s) decoding:\n\n* 'Since' expected type 'int64', got unconvertible type 'string'")
	assert.Equal(t, segment, Segment{})
}

func TestGetSegmentReturnsGetAllChangesError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer testServer.Close()
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	segment, err := result.getSegment("mock-segment-name")

	// Validate that GetSegment function returns GetAllChanges error
	assert.EqualError(t, err, "Non-OK HTTP status: 401 Unauthorized")
	assert.Equal(t, segment, Segment{})
}

func TestGetSegmentsForSplitsReturnsGetSplitError(t *testing.T) {
	// Arrange
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"since":"wrong-type", "till":10}`)
	}))
	defer testServer.Close()
	conditions := []dtos.ConditionDTO{}
	json.Unmarshal([]byte(mockConditions), &conditions)
	split := dtos.SplitDTO{Conditions: conditions}
	splits := []dtos.SplitDTO{split}
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	segments, usingSegmentsCount, err := result.GetSegmentsForSplits(splits)

	// Validate that GetSegmentForSplits function returns error from GetSegment
	assert.EqualError(t, err, "error when decode data to segment: 1 error(s) decoding:\n\n* 'Since' expected type 'int64', got unconvertible type 'string'")
	assert.Equal(t, segments, []Segment{})
	assert.Equal(t, usingSegmentsCount, 0)
}

func TestGetSegmentsForSplitsReturnsValid(t *testing.T) {
	// Arrange
	handler := &mockHandler{}
	testServer := httptest.NewServer(handler)
	defer testServer.Close()
	conditions := []dtos.ConditionDTO{}
	json.Unmarshal([]byte(mockConditions), &conditions)
	split := dtos.SplitDTO{Conditions: conditions}
	splits := []dtos.SplitDTO{split}
	result := NewSplitioAPIBinding(mockSplitioAPIKey, testServer.URL)

	// Act
	segments, usingSegmentsCount, err := result.GetSegmentsForSplits(splits)

	// Validate that GetSegmentForSplits function returns correct segments
	assert.Nil(t, err)
	assert.Equal(t, len(segments), 1)
	assert.Equal(t, len(segments[0].Added), 4)
	assert.Equal(t, segments[0].Name, "mock-segment")
	assert.Equal(t, usingSegmentsCount, 1)
}
