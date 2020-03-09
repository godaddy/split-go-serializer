package serializer

import (
	"testing"

	"github.com/godaddy/split-go-serializer/v2/poller"
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

func (fetcher *mockFetcher) GetSerializedData() string {
	return "foo"
}

func (fetcher *mockFetcher) GetSerializedDataSubset([]string) string {
	return "bar"
}

func (fetcher *mockFetcher) GetSplitData() poller.SplitData {
	if !fetcher.hasData {
		return poller.SplitData{}
	}

	mockSplits := map[string]dtos.SplitDTO{
		"mock-split-1": {
			Name:   "mock-split-1",
			Status: "mock-status-1",
		},
	}
	mockSegments := map[string]dtos.SegmentChangesDTO{
		"mock-segment-1": {
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
	result, err := serializer.GetSerializedData([]string{})

	// Validate that returned logging script contains the string from poller cache
	expectedLoggingScript := "foo"
	assert.Equal(t, result, expectedLoggingScript)
	assert.Nil(t, err)
}

func TestGetSerializedDataWithNonEmptySplits(t *testing.T) {
	// Arrange
	splits := []string{"test-experiment-1"}
	serializer := NewSerializer(&mockFetcher{hasData: true})

	// Act
	result, err := serializer.GetSerializedData(splits)

	// Validate that returned logging script contains the string from poller cache
	expectedLoggingScript := "bar"
	assert.Equal(t, result, expectedLoggingScript)
	assert.Nil(t, err)
}
