package poller

import (
	"errors"
	"testing"
	"time"

	"github.com/godaddy/split-go-serializer/api"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

func TestNewPollerValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	// Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that returned Poller has the correct type and values
	assert.Equal(t, result.pollingRateSeconds, pollingRateSeconds)
	assert.Equal(t, result.serializeSegments, serializeSegments)
	assert.IsType(t, result.splitioAPIBinding, api.SplitioAPIBinding{})
	assert.Equal(t, result.Cache, 0)
}

func TestNewSerializerDefaultPollingRateSeconds(t *testing.T) {
	// Arrange
	pollingRateSeconds := 0

	// Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)
	expectedPollingRateSeconds := 300

	// Validate that returned Poller has the correct type and values
	assert.Equal(t, result.pollingRateSeconds, expectedPollingRateSeconds)
}

func TestPollforChangesValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)
	result.pollForChanges()

	// Validate that after calling PollforChanges it returns the right value
	assert.Equal(t, 1, result.Cache)
	assert.Nil(t, result.Error)
}

func TestStartValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that after calling Start the cache is updated
	assert.Equal(t, result.Cache, 0)
	result.Start()
	time.Sleep(2 * time.Second)
	assert.True(t, result.Cache > 1)
	result.quit <- true
}

func TestJobsUpdatesCache(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that after calling jobs the cache is updated
	assert.Equal(t, result.Cache, 0)
	time.AfterFunc(3*time.Second, func() {
		result.quit <- true
	})
	result.jobs()
	assert.True(t, result.Cache > 0)
}

func TestJobsStopsWhenError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that Jobs stop if error is received
	assert.Equal(t, result.Cache, 0)
	time.AfterFunc(3*time.Second, func() {
		result.errorChannel <- errors.New("mock error")
	})
	result.jobs()
	assert.True(t, result.Cache > 0)
}

func TestStopValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that when Stop is called, jobs will stop
	assert.Equal(t, result.Cache, 0)
	time.AfterFunc(3*time.Second, func() { result.Stop() })
	// jobs is an infinite loop, and expect to stop after Stop is called
	result.jobs()
	assert.True(t, result.Cache > 0)
}
