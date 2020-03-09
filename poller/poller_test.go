package poller

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/godaddy/split-go-serializer/v2/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

type mockSplitio struct {
	mockSince              int64
	mockUsingSegmentsCount int
	getSplitValid          bool
	getSegmentValid        bool
}

func (splitio *mockSplitio) GetSplits() (map[string]dtos.SplitDTO, int64, error) {
	if splitio.getSplitValid {
		mockSplit := dtos.SplitDTO{Name: "mock-split"}
		mockSplitMap := map[string]dtos.SplitDTO{
			"mock-split": mockSplit,
		}
		splitio.mockSince++
		return mockSplitMap, splitio.mockSince, nil
	}
	return nil, 0, fmt.Errorf("Error from splitio API when getting splits")
}

func (splitio *mockSplitio) GetSegmentsForSplits(splits map[string]dtos.SplitDTO) (map[string]dtos.SegmentChangesDTO, int, error) {
	if splitio.getSegmentValid {
		mockSegment := dtos.SegmentChangesDTO{
			Name: "mock-segment",
		}
		mockSegmentMap := map[string]dtos.SegmentChangesDTO{
			"mock-segment": mockSegment,
		}
		splitio.mockUsingSegmentsCount++
		return mockSegmentMap, splitio.mockUsingSegmentsCount, nil
	}
	return nil, 0, fmt.Errorf("Error from splitio API when getting segments")
}

// getSplitData helper returns split data cache results
func getSplitData(poller *Poller) SplitData {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).SplitData
}

func TestNewPollerValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	// Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments, nil)

	// Validate that returned Poller has the correct type and values
	assert.Equal(t, result.pollingRateSeconds, pollingRateSeconds)
	assert.Equal(t, result.serializeSegments, serializeSegments)
	assert.IsType(t, result.splitio, &api.SplitioAPIBinding{})
}

func TestNewSerializerDefaultPollingRateSeconds(t *testing.T) {
	// Arrange
	pollingRateSeconds := 0

	// Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments, nil)
	expectedPollingRateSeconds := 300

	// Validate that returned Poller has the correct type and values
	assert.Equal(t, result.pollingRateSeconds, expectedPollingRateSeconds)
}

func TestPollforChangesValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})
	result.pollForChanges()
	returnedCache := getSplitData(result)

	// Validate that after calling PollforChanges it returns the right value
	assert.Equal(t, int64(1), returnedCache.Since)
	assert.Equal(t, 1, returnedCache.UsingSegmentsCount)
}

func TestStartValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Validate that after calling Start the cache is updated
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	result.Start()
	time.Sleep(2 * time.Second)
	cacheAfterStart := getSplitData(result)
	assert.True(t, cacheAfterStart.Since > 1)
	assert.True(t, cacheAfterStart.UsingSegmentsCount > 0)
	result.quit <- true
}

func TestStopValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, false,
		&mockSplitio{getSplitValid: true})

	// Validate that when Stop is called, jobs will stop
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	result.Stop()
	cacheAfterStop := getSplitData(result)
	assert.True(t, cacheAfterStop.Since > 0)
	time.Sleep(2 * time.Second)
	assert.Equal(t, cacheAfterStop.Since, getSplitData(result).Since)
}

func TestJobsUpdatesCache(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Validate that after calling jobs the cache is updated
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	time.Sleep(2 * time.Second)
	cacheAfterStart := getSplitData(result)
	assert.True(t, cacheAfterStart.Since > 0)
	assert.True(t, cacheAfterStart.UsingSegmentsCount > 0)
	result.quit <- true
}

func TestJobsStopsWhenQuit(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, false,
		&mockSplitio{getSplitValid: true})

	// Validate that Jobs stop if quit is set to true
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	assert.True(t, getSplitData(result).Since > 0)
	result.quit <- true
	cacheAfterStop := getSplitData(result)
	time.Sleep(2 * time.Second)
	assert.Equal(t, cacheAfterStop.Since, getSplitData(result).Since)
}

func TestJobsCanRunTwiceAfterStop(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Validate that jobs can be run more than once

	// First loop
	cacheBeforeStart := getSplitData(result)
	serializedCacheBeforeStart := result.GetSerializedData()
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	assert.Equal(t, serializedCacheBeforeStart, emptyCacheLoggingScript)
	go result.jobs()
	time.Sleep(3 * time.Second)

	// assert loop calls function so cache is updated
	cacheAfterStart := getSplitData(result)
	serializedCacheAfterStart := result.GetSerializedData()
	assert.True(t, cacheAfterStart.Since > 0)
	assert.True(t, cacheAfterStart.UsingSegmentsCount > 0)
	assert.Equal(t, cacheAfterStart.Splits["mock-split"].Name, "mock-split")
	assert.Equal(t, cacheAfterStart.Segments["mock-segment"].Name, "mock-segment")
	expectedSerializedScript := generateSerializedData(cacheAfterStart)
	assert.Equal(t, serializedCacheAfterStart, expectedSerializedScript)
	result.Stop()

	firstSince := getSplitData(result).Since
	firstCount := getSplitData(result).UsingSegmentsCount
	time.Sleep(3 * time.Second)
	// verfify Cache didn't update after stop
	assert.Equal(t, getSplitData(result).Since, firstSince)
	assert.Equal(t, getSplitData(result).UsingSegmentsCount, firstCount)

	// Second loop
	go result.jobs()
	time.Sleep(2 * time.Second)

	// verfify cache is updated due to second loop
	assert.True(t, getSplitData(result).Since > firstSince)
	assert.True(t, getSplitData(result).UsingSegmentsCount > firstCount)
	result.Stop()
}

func TestPollforChangesReturnsGetSplitsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: false, getSegmentValid: false})
	hasErr := false
	var err error

	// Validate that error is received when getSplits returns error and cache isn't updated
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := getSplitData(result)
	assert.Equal(t, cacheAfterError, SplitData{})
	assert.Equal(t, cacheAfterError.Since, int64(0))
	assert.Equal(t, cacheAfterError.UsingSegmentsCount, 0)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting splits")
	result.Stop()
}

func TestPollforChangesReturnsGetSegmentsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: false})
	hasErr := false
	var err error

	// Validate that error is received when getSegments returns error and cache isn't updated
	cacheBeforeStart := getSplitData(result)
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := getSplitData(result)
	assert.Equal(t, cacheAfterError, SplitData{})
	assert.Equal(t, cacheAfterError.Since, int64(0))
	assert.Equal(t, cacheAfterError.UsingSegmentsCount, 0)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting segments")
	result.Stop()
}

func TestJobsKeepRunningAfterGettingError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1
	mockSplitioDataGetter := &mockSplitio{
		getSplitValid: false,
	}

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		mockSplitioDataGetter)
	hasErr := false
	var err error

	// Validate that after first time error cache can still be updated

	// first loop
	cacheBeforeStart := getSplitData(result)
	serializedCacheBeforeStart := result.GetSerializedData()
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	assert.Equal(t, serializedCacheBeforeStart, emptyCacheLoggingScript)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := getSplitData(result)
	serializedCacheAfterError := result.GetSerializedData()
	assert.Equal(t, cacheAfterError, SplitData{})
	assert.Equal(t, cacheAfterError.Since, int64(0))
	assert.Equal(t, cacheAfterError.UsingSegmentsCount, 0)
	assert.Equal(t, serializedCacheAfterError, emptyCacheLoggingScript)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting splits")

	// after setting getSplit, getSegment to true, jobs is still running and cache is updated
	mockSplitioDataGetter.getSplitValid = true
	mockSplitioDataGetter.getSegmentValid = true
	time.Sleep(5 * time.Second)
	cacheSecondRound := getSplitData(result)
	serializedCacheSecondRound := result.GetSerializedData()
	assert.True(t, cacheSecondRound.Since > 0)
	assert.True(t, cacheSecondRound.UsingSegmentsCount > 0)
	assert.Equal(t, cacheSecondRound.Splits["mock-split"].Name, "mock-split")
	assert.Equal(t, cacheSecondRound.Segments["mock-segment"].Name, "mock-segment")
	expectedSerializedScript := generateSerializedData(cacheSecondRound)
	assert.Equal(t, serializedCacheSecondRound, expectedSerializedScript)
	result.Stop()
}

func TestGenerateSerializedDataValid(t *testing.T) {
	// Arrange
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
	mockSplitData := SplitData{
		Splits:             mockSplits,
		Since:              1,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}
	// Act
	result := generateSerializedData(mockSplitData)

	// Validate that returned logging script contains a valid SplitData
	stringSplits := `{"mock-split-1":"{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-1\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-1\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}"}`
	stringSegments := `{"mock-segment-1":"{\"name\":\"mock-segment-1\",\"added\":[\"foo\",\"bar\"],\"removed\":null,\"since\":20,\"till\":20}"}`
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, stringSplits, 1, stringSegments, 2)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataMarshalEmptyCache(t *testing.T) {
	// Act
	result := generateSerializedData(SplitData{})

	// Validate that returned logging script contains a valid SplitData
	expectedLoggingScript := fmt.Sprintf(emptyCacheLoggingScript)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataSplitError(t *testing.T) {
	// Act
	result := generateSerializedData(SplitData{})

	// Validate that returned logging script contains a valid SplitData
	expectedLoggingScript := fmt.Sprintf(emptyCacheLoggingScript)
	assert.Equal(t, result, expectedLoggingScript)
}
