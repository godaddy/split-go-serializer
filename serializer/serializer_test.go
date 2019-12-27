package serializer

import (
	"fmt"
	"testing"

	"github.com/godaddy/split-go-serializer/poller"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

type mockFetcher struct {
	hasData bool
}

func (fetcher *mockFetcher) Start() {
}

func (fetcher *mockFetcher) Stop() {
}

func (fetcher *mockFetcher) GetSplitData() poller.SplitData {
	if !fetcher.hasData {
		return poller.SplitData{}
	}

	mockSplits := []dtos.SplitDTO{
		{
			Name:   "mock-split-1",
			Status: "mock-status-1",
		},
	}
	mockSegments := []dtos.SegmentChangesDTO{
		{
			Name:  "mock-segment-1",
			Added: []string{"foo", "bar"},
			Since: 20,
			Till:  20,
		},
	}
	testCache := poller.SplitData{
		Splits:             mockSplits,
		Since:              1,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}
	return testCache

}

func TestNewSerializerValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400
	testPoller := poller.NewPoller(testKey, pollingRateSeconds, serializeSegments, nil)

	// Act
	result := NewSerializer(testPoller)

	// Validate that returned Serializer has the correct type and values
	assert.IsType(t, result.poller, &poller.Poller{})
}

func TestGetSerializedDataValid(t *testing.T) {
	// Arrange
	serializer := NewSerializer(&mockFetcher{hasData: true})

	// Act
	result, err := serializer.GetSerializedData()

	// Validate that returned logging script contains a valid Cache data
	expectedData := "{\"Splits\":[{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-1\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-1\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}],\"Since\":1,\"Segments\":[{\"name\":\"mock-segment-1\",\"added\":[\"foo\",\"bar\"],\"removed\":null,\"since\":20,\"till\":20}],\"UsingSegmentsCount\":2}"
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, expectedData)
	assert.Equal(t, result, expectedLoggingScript)
	assert.Nil(t, err)
}

func TestGetSerializedDataMarshalEmptyCache(t *testing.T) {
	// Arrange
	serializer := NewSerializer(&mockFetcher{hasData: false})

	// Act
	result, err := serializer.GetSerializedData()

	// Validate that returned logging script contains a valid Cache data
	expectedData := "{\"Splits\":null,\"Since\":0,\"Segments\":null,\"UsingSegmentsCount\":0}"
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, expectedData)
	assert.Equal(t, result, expectedLoggingScript)
	assert.Nil(t, err)
}
