package poller

import (
	"fmt"
	"testing"
	"time"

	"github.com/godaddy/split-go-serializer/api"
	"github.com/splitio/go-client/splitio/service/dtos"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

type mockSplitioDataGetter struct {
	mockSince              int64
	mockUsingSegmentsCount int
	getSplitValid          bool
	getSegmentValid        bool
}

func (splitioDataGetter *mockSplitioDataGetter) GetSplits() ([]dtos.SplitDTO, int64, error) {
	if splitioDataGetter.getSplitValid {
		fmt.Println("return mock split in mockGetter", time.Now())
		mockSplit := dtos.SplitDTO{Name: "mock-split"}
		splitioDataGetter.mockSince++
		return []dtos.SplitDTO{mockSplit}, splitioDataGetter.mockSince, nil
	}
	return nil, 0, fmt.Errorf("Error from splitio API when getting splits")
}

func (splitioDataGetter *mockSplitioDataGetter) GetSegmentsForSplits(splits []dtos.SplitDTO) ([]dtos.SegmentChangesDTO, int, error) {
	if splitioDataGetter.getSegmentValid {
		fmt.Println("return mock segment in mockGetter", time.Now())
		mockSegment := dtos.SegmentChangesDTO{
			Name: "mock-segment",
		}
		splitioDataGetter.mockUsingSegmentsCount++
		return []dtos.SegmentChangesDTO{mockSegment}, splitioDataGetter.mockUsingSegmentsCount, nil
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
	assert.IsType(t, result.splitioDataGetter, &api.SplitioAPIBinding{})
	assert.Equal(t, result.Cache, Cache{})
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
		&mockSplitioDataGetter{getSplitValid: true, getSegmentValid: true})
	result.pollForChanges()

	// Validate that after calling PollforChanges it returns the right value
	assert.Equal(t, int64(1), result.Cache.Since)
	assert.Equal(t, 1, result.Cache.UsingSegmentsCount)
}

func TestStartValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitioDataGetter{getSplitValid: true, getSegmentValid: true})

	// Validate that after calling Start the cache is updated
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	result.Start()
	time.Sleep(2 * time.Second)
	assert.True(t, result.Cache.Since > 1)
	assert.True(t, result.Cache.UsingSegmentsCount > 0)
	result.quit <- true
}

func TestStopValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, false,
		&mockSplitioDataGetter{getSplitValid: true})

	// Validate that when Stop is called, jobs will stop
	assert.Equal(t, result.Cache.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	result.Stop()
	sinceAfterStop := result.Cache.Since
	assert.True(t, sinceAfterStop > 0)
	time.Sleep(2 * time.Second)
	assert.Equal(t, sinceAfterStop, result.Cache.Since)
}

func TestJobsUpdatesCache(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitioDataGetter{getSplitValid: true, getSegmentValid: true})

	// Validate that after calling jobs the cache is updated
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	go result.jobs()
	time.Sleep(2 * time.Second)
	assert.True(t, result.Cache.Since > 0)
	assert.True(t, result.Cache.UsingSegmentsCount > 0)
	result.quit <- true
}

func TestJobsStopsWhenQuit(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, false,
		&mockSplitioDataGetter{getSplitValid: true})

	// Validate that Jobs stop if quit is set to true
	assert.Equal(t, result.Cache.Since, int64(0))
	go result.jobs()
	time.Sleep(2 * time.Second)
	assert.True(t, result.Cache.Since > 0)
	result.quit <- true
	sinceAfterStop := result.Cache.Since
	time.Sleep(2 * time.Second)
	assert.Equal(t, sinceAfterStop, result.Cache.Since)
}

func TestJobsCanRunTwiceAfterStop(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitioDataGetter{getSplitValid: true, getSegmentValid: true})

	// Validate that jobs can be run more than once

	// First loop
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	go result.jobs()
	time.Sleep(3 * time.Second)

	// assert loop calls function so cache is updated
	assert.True(t, result.Cache.Since > 0)
	assert.True(t, result.Cache.UsingSegmentsCount > 0)
	assert.Equal(t, result.Cache.Splits[0].Name, "mock-split")
	assert.Equal(t, result.Cache.Segments[0].Name, "mock-segment")
	result.Stop()

	firstSince := result.Cache.Since
	firstCount := result.Cache.UsingSegmentsCount
	time.Sleep(3 * time.Second)
	// verfify Cache didn't update after stop
	assert.Equal(t, result.Cache.Since, firstSince)
	assert.Equal(t, result.Cache.UsingSegmentsCount, firstCount)

	// Second loop
	go result.jobs()
	time.Sleep(2 * time.Second)

	// verfify cache is updated due to second loop
	assert.True(t, result.Cache.Since > firstSince)
	assert.True(t, result.Cache.UsingSegmentsCount > firstCount)
	result.Stop()
}

func TestPollforChangesReturnsGetSplitsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitioDataGetter{getSplitValid: false, getSegmentValid: false})
	hasErr := false
	var err error

	// Validate that error is received when getSplits returns error and cache isn't updated
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting splits")
	result.Stop()
}

func TestPollforChangesReturnsGetSegmentsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		&mockSplitioDataGetter{getSplitValid: true, getSegmentValid: false})
	hasErr := false
	var err error

	// Validate that error is received when getSegments returns error and cache isn't updated
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting segments")
	result.Stop()
}

func TestJobsKeepRunningAfterGettingError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1
	mockSplitioDataGetter := &mockSplitioDataGetter{
		getSplitValid: false,
	}

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments,
		mockSplitioDataGetter)
	hasErr := false
	var err error

	// Validate that after first time error cache can still be updated

	// first loop
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	go result.jobs()
	err = <-result.Error
	if err != nil {
		hasErr = true
	}
	assert.Equal(t, result.Cache, Cache{})
	assert.Equal(t, result.Cache.Since, int64(0))
	assert.Equal(t, result.Cache.UsingSegmentsCount, 0)
	assert.True(t, hasErr)
	assert.EqualError(t, err, "Error from splitio API when getting splits")

	// after setting getSplit, getSegment to true, jobs is still running and cache is updated
	mockSplitioDataGetter.getSplitValid = true
	mockSplitioDataGetter.getSegmentValid = true
	time.Sleep(5 * time.Second)
	assert.True(t, result.Cache.Since > 0)
	assert.True(t, result.Cache.UsingSegmentsCount > 0)
	assert.Equal(t, result.Cache.Splits[0].Name, "mock-split")
	assert.Equal(t, result.Cache.Segments[0].Name, "mock-segment")
	result.Stop()
}
