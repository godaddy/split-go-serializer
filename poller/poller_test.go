package poller

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/godaddy/split-go-serializer/v3/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/stretchr/testify/assert"
)

const (
	testKey                   = "someKey"
	serializeSegments         = true
	stringSegments            = `{"mock-segment-1":"{\"name\":\"mock-segment-1\",\"added\":[\"foo\",\"bar\"],\"removed\":null,\"since\":20,\"till\":20}"}`
	stringSegmentsMockSplitIo = `{"mock-segment":"{\"name\":\"mock-segment\",\"added\":null,\"removed\":null,\"since\":0,\"till\":0}"}`
)

var mockMultipleSplits = map[string]dtos.SplitDTO{
	"mock-split-1": {
		Name:   "mock-split-1",
		Status: "mock-status-1",
	},
	"mock-split-2": {
		Name:   "mock-split-2",
		Status: "mock-status-2",
	},
	"mock-split-3": {
		Name:   "mock-split-3",
		Status: "mock-status-3",
	},
}

var mockSegments = map[string]dtos.SegmentChangesDTO{
	"mock-segment-1": {
		Name:  "mock-segment-1",
		Added: []string{"foo", "bar"},
		Since: 20,
		Till:  20,
	},
}

type mockSplitio struct {
	mockSince              int64
	mockUsingSegmentsCount int
	getSplitValid          bool
	getSegmentValid        bool
	deterministic          bool
}

func (splitio *mockSplitio) GetSplits() (map[string]dtos.SplitDTO, int64, error) {
	if splitio.getSplitValid {
		mockSplit := dtos.SplitDTO{Name: "mock-split"}
		mockSplitMap := map[string]dtos.SplitDTO{
			"mock-split":   mockSplit,
			"mock-split-2": dtos.SplitDTO{Name: "mock-split-2"},
			"mock-split-3": dtos.SplitDTO{Name: "mock-split-3"},
		}
		if !splitio.deterministic {
			splitio.mockSince++
		}
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
		if !splitio.deterministic {
			splitio.mockUsingSegmentsCount++
		}
		return mockSegmentMap, splitio.mockUsingSegmentsCount, nil
	}
	return nil, 0, fmt.Errorf("Error from splitio API when getting segments")
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
	returnedCache := result.getSplitData()

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
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	result.Start()
	time.Sleep(2 * time.Second)
	cacheAfterStart := result.getSplitData()
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
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	result.Stop()
	cacheAfterStop := result.getSplitData()
	assert.True(t, cacheAfterStop.Since > 0)
	time.Sleep(2 * time.Second)
	assert.Equal(t, cacheAfterStop.Since, result.getSplitData().Since)
}

func TestJobsUpdatesCache(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Validate that after calling jobs the cache is updated
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	time.Sleep(2 * time.Second)
	cacheAfterStart := result.getSplitData()
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
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	assert.True(t, result.getSplitData().Since > 0)
	result.quit <- true
	cacheAfterStop := result.getSplitData()
	time.Sleep(2 * time.Second)
	assert.Equal(t, cacheAfterStop.Since, result.getSplitData().Since)
}

func TestJobsCanRunTwiceAfterStop(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Validate that jobs can be run more than once

	// First loop
	cacheBeforeStart := result.getSplitData()
	serializedCacheBeforeStart := result.GetSerializedData([]string{})
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	assert.Equal(t, serializedCacheBeforeStart, emptyCacheLoggingScript)
	go result.jobs()
	time.Sleep(3 * time.Second)

	// assert loop calls function so cache is updated
	cacheAfterStart := result.getSplitData()
	serializedCacheAfterStart := result.GetSerializedData([]string{})
	assert.True(t, cacheAfterStart.Since > 0)
	assert.True(t, cacheAfterStart.UsingSegmentsCount > 0)
	assert.Equal(t, cacheAfterStart.Splits["mock-split"].Name, "mock-split")
	assert.Equal(t, cacheAfterStart.Segments["mock-segment"].Name, "mock-segment")
	expectedSerializedScript := result.generateSerializedData(cacheAfterStart, []string{})
	assert.Equal(t, serializedCacheAfterStart, expectedSerializedScript)
	result.Stop()

	firstSince := result.getSplitData().Since
	firstCount := result.getSplitData().UsingSegmentsCount
	time.Sleep(3 * time.Second)
	// verfify Cache didn't update after stop
	assert.Equal(t, result.getSplitData().Since, firstSince)
	assert.Equal(t, result.getSplitData().UsingSegmentsCount, firstCount)

	// Second loop
	go result.jobs()
	time.Sleep(2 * time.Second)

	// verfify cache is updated due to second loop
	assert.True(t, result.getSplitData().Since > firstSince)
	assert.True(t, result.getSplitData().UsingSegmentsCount > firstCount)
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
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := result.getSplitData()
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
	cacheBeforeStart := result.getSplitData()
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := result.getSplitData()
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
	cacheBeforeStart := result.getSplitData()
	serializedCacheBeforeStart := result.GetSerializedData([]string{})
	assert.Equal(t, cacheBeforeStart, SplitData{})
	assert.Equal(t, cacheBeforeStart.Since, int64(0))
	assert.Equal(t, cacheBeforeStart.UsingSegmentsCount, 0)
	assert.Equal(t, serializedCacheBeforeStart, emptyCacheLoggingScript)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	cacheAfterError := result.getSplitData()
	serializedCacheAfterError := result.GetSerializedData([]string{})
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
	cacheSecondRound := result.getSplitData()
	serializedCacheSecondRound := result.GetSerializedData([]string{})
	assert.True(t, cacheSecondRound.Since > 0)
	assert.True(t, cacheSecondRound.UsingSegmentsCount > 0)
	assert.Equal(t, cacheSecondRound.Splits["mock-split"].Name, "mock-split")
	assert.Equal(t, cacheSecondRound.Segments["mock-segment"].Name, "mock-segment")
	expectedSerializedScript := result.generateSerializedData(cacheSecondRound, []string{})
	assert.Equal(t, serializedCacheSecondRound, expectedSerializedScript)
	result.Stop()
}

func TestGetSerializedDataWithSplitNamesPassedIn(t *testing.T) {
	// Arrange
	splitNames := []string{"mock-split-2"}
	pollingRateSeconds := 1
	deterministic := true

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitio{mockSince: 10, mockUsingSegmentsCount: 10, getSplitValid: true, getSegmentValid: true, deterministic: deterministic})

	// Validate that GetSerializedData returns serialized data subset properly

	// before start, cached serialized subsets should be an empty logging script for the subset and the serialized data returned should be an empty logging script
	serializedCachedDataSubsetsBeforeStart := result.getCachedSerializedDataSubsets()
	subsetBeforeStart := result.GetSerializedData(splitNames)
	assert.Equal(t, serializedCachedDataSubsetsBeforeStart, map[string]string{
		"mock-split-2": emptyCacheLoggingScript,
	})
	assert.Equal(t, subsetBeforeStart, emptyCacheLoggingScript)

	result.Start()
	time.Sleep(2 * time.Second)

	// after starting, cached serialized subsets should contain a valid logging script
	cacheSplitData := result.getSplitData()
	serializedCachedDataSubsetsAfterStart := result.getCachedSerializedDataSubsets()
	subsetAfterStart := result.GetSerializedData(splitNames)
	expectedSerializedScript := result.generateSerializedData(cacheSplitData, splitNames)
	assert.Equal(t, serializedCachedDataSubsetsAfterStart, map[string]string{
		"mock-split-2": expectedSerializedScript,
	})
	assert.Equal(t, subsetAfterStart, expectedSerializedScript)
	result.quit <- true
}

func TestGetUpdatedSerializedDataSubsetsValid(t *testing.T) {
	// Arrange
	deterministic := true
	mockSince := int64(1)
	mockUsingSegmentsCount := 3
	mockSplitData := SplitData{
		Splits:             mockMultipleSplits,
		Since:              mockSince,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}
	serializedDataSubsets := map[string]string{
		"mock-split-1.mock-split-2":              "",
		"mock-split-1.mock-split-2.mock-split-3": "",
		"mock-split-2":                           "",
	}
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{mockUsingSegmentsCount: mockUsingSegmentsCount, getSplitValid: true, getSegmentValid: true, deterministic: deterministic})
	cache := Cache{
		splitData:             mockSplitData,
		serializedData:        poller.GetSerializedData([]string{}),
		serializedDataSubsets: serializedDataSubsets,
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&cache))

	// Act
	result := poller.getUpdatedSerializedDataSubsets(mockSplitData)

	// Validate that an updated serializedDataSubsets, with correct logging scripts, is returned
	stringSplit := `"mock-split-%v":"{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-%v\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-%v\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}"`
	mockSplitOneString := fmt.Sprintf(stringSplit, 1, 1, 1)
	mockSplitTwoString := fmt.Sprintf(stringSplit, 2, 2, 2)
	mockSplitThreeString := fmt.Sprintf(stringSplit, 3, 3, 3)
	firstSplitDataString := fmt.Sprintf(`{%v,%v}`, mockSplitOneString, mockSplitTwoString)
	secondSplitDataString := fmt.Sprintf(`{%v,%v,%v}`, mockSplitOneString, mockSplitTwoString, mockSplitThreeString)
	thirdSplitDataString := fmt.Sprintf(`{%v}`, mockSplitTwoString)

	expectedUpdatedSerializedDataSubsets := map[string]string{
		"mock-split-1.mock-split-2":              fmt.Sprintf(formattedLoggingScript, firstSplitDataString, mockSince, stringSegmentsMockSplitIo, mockUsingSegmentsCount),
		"mock-split-1.mock-split-2.mock-split-3": fmt.Sprintf(formattedLoggingScript, secondSplitDataString, mockSince, stringSegmentsMockSplitIo, mockUsingSegmentsCount),
		"mock-split-2":                           fmt.Sprintf(formattedLoggingScript, thirdSplitDataString, mockSince, stringSegmentsMockSplitIo, mockUsingSegmentsCount),
	}
	assert.Equal(t, result, expectedUpdatedSerializedDataSubsets)
}
func TestGenerateSerializedDataValid(t *testing.T) {
	// Arrange
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})
	mockSplits := map[string]dtos.SplitDTO{
		"mock-split-1": {
			Name:   "mock-split-1",
			Status: "mock-status-1",
		},
	}
	mockSplitData := SplitData{
		Splits:             mockSplits,
		Since:              1,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}
	// Act
	result := poller.generateSerializedData(mockSplitData, []string{})

	// Validate that returned logging script contains a valid SplitData
	stringSplits := `{"mock-split-1":"{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-1\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-1\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}"}`
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, stringSplits, 1, stringSegments, 2)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataWithNonEmptySplitNames(t *testing.T) {
	// Arrange
	mockSince := int64(1)
	mockUsingSegmentsCount := 200
	deterministic := true
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{mockUsingSegmentsCount: mockUsingSegmentsCount, getSplitValid: true, getSegmentValid: true, deterministic: deterministic})
	splitNames := []string{"mock-split-2"}
	mockSplitData := SplitData{
		Splits:             mockMultipleSplits,
		Since:              mockSince,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}

	// Act
	result := poller.generateSerializedData(mockSplitData, splitNames)

	// Validate that returned logging script only contains SplitData for splits passed in,
	// that segments data is from the mocked GetSegmentsForSplits response,
	// and that the usingSegmentsCount is our mocked value
	stringSplits := `{"mock-split-2":"{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-2\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-2\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}"}`
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, stringSplits, mockSince, stringSegmentsMockSplitIo, mockUsingSegmentsCount)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataWithNonEmptySplitNamesAndFalseSerializeSegments(t *testing.T) {
	// Arrange
	getSegments := false
	emptySegmentsData := map[string]dtos.SegmentChangesDTO{}
	zeroUsingSegmentsCount := 0

	mockSince := int64(1)
	poller := NewPoller(testKey, 1, getSegments,
		&mockSplitio{mockUsingSegmentsCount: 200, getSplitValid: true, getSegmentValid: true, deterministic: true})
	splitNames := []string{"mock-split-2"}
	mockSplitData := SplitData{
		Splits:             mockMultipleSplits,
		Since:              mockSince,
		Segments:           emptySegmentsData,
		UsingSegmentsCount: zeroUsingSegmentsCount,
	}

	// Act
	result := poller.generateSerializedData(mockSplitData, splitNames)

	// Validate that returned logging script only contains SplitData for splits passed in,
	// and that there is an empty segmentsData and zero usingSegmentsCount
	stringSplits := `{"mock-split-2":"{\"changeNumber\":0,\"trafficTypeName\":\"\",\"name\":\"mock-split-2\",\"trafficAllocation\":0,\"trafficAllocationSeed\":0,\"seed\":0,\"status\":\"mock-status-2\",\"killed\":false,\"defaultTreatment\":\"\",\"algo\":0,\"conditions\":null,\"configurations\":null}"}`
	expectedSegmentsData := "{}"
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, stringSplits, mockSince, expectedSegmentsData, zeroUsingSegmentsCount)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataWithNonEmptySplitNamesAndInvalidSegmentsReturnsEmptyScript(t *testing.T) {
	// Arrange
	getSegmentValid := false
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: getSegmentValid})
	splitNames := []string{"mock-split-2"}
	mockSplitData := SplitData{
		Splits:             mockMultipleSplits,
		Since:              1,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}

	// Act
	result := poller.generateSerializedData(mockSplitData, splitNames)

	// Validate that output is an empty logging script
	assert.Equal(t, result, emptyCacheLoggingScript)
}

func TestGenerateSerializedDataWithInvalidSplitsReturnsNoSplitsData(t *testing.T) {
	// Arrange
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})
	splitNames := []string{"invalid-split-1", "invalid-split-2"}
	mockSplitData := SplitData{
		Splits:             mockMultipleSplits,
		Since:              1,
		Segments:           mockSegments,
		UsingSegmentsCount: 2,
	}

	// Act
	result := poller.generateSerializedData(mockSplitData, splitNames)

	// Validate that returned logging script does not contain any splits data
	emptySplits := "{}"
	expectedLoggingScript := fmt.Sprintf(formattedLoggingScript, emptySplits, 1, stringSegmentsMockSplitIo, 1)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataMarshalEmptyCache(t *testing.T) {
	// Arrange
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Act
	result := poller.generateSerializedData(SplitData{}, []string{})

	// Validate that returned logging script contains a valid SplitData
	expectedLoggingScript := fmt.Sprintf(emptyCacheLoggingScript)
	assert.Equal(t, result, expectedLoggingScript)
}

func TestGenerateSerializedDataSplitError(t *testing.T) {
	// Arrange
	poller := NewPoller(testKey, 1, serializeSegments,
		&mockSplitio{getSplitValid: true, getSegmentValid: true})

	// Act
	result := poller.generateSerializedData(SplitData{}, []string{})

	// Validate that returned logging script contains a valid SplitData
	expectedLoggingScript := fmt.Sprintf(emptyCacheLoggingScript)
	assert.Equal(t, result, expectedLoggingScript)
}
