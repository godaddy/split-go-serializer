package poller

import (
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
	assert.Equal(t, result.PollingRateSeconds, pollingRateSeconds)
	assert.Equal(t, result.SerializeSegments, serializeSegments)
	assert.Equal(t, result.Cache, "")
	assert.IsType(t, result.SplitioAPIBinding, api.SplitioAPIBinding{})
}

func TestNewSerializerDefaultPollingRateSeconds(t *testing.T) {
	// Arrange
	pollingRateSeconds := 0

	// Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)
	expectedPollingRateSeconds := 300

	// Validate that returned Poller has the correct type and values
	assert.Equal(t, result.PollingRateSeconds, expectedPollingRateSeconds)
}

func TestPollforChangesValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)
	err := result.pollForChanges()

	// Validate that after calling PollforChanges it returns the right value
	assert.Nil(t, err)
	assert.Equal(t, "data from splitChanges and segmentChanges", result.Cache)
}

func TestPollValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that after calling Poll the cache is updated
	assert.Equal(t, result.Cache, "")
	result.Poll()
	assert.Equal(t, result.Cache, "data from splitChanges and segmentChanges")
}

func TestJobsValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that jobs function triggers pollForChanges and updates the cache
	assert.Equal(t, result.Cache, "")
	// time.AfterFunc(3*time.Second, func() { result.Stop() })
	go result.jobs()
	time.Sleep(2 * time.Second)
	result.quit <- true
	assert.Equal(t, result.Cache, "data from splitChanges and segmentChanges")
}

func TestStopValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 1

	//Act
	result := NewPoller(testKey, pollingRateSeconds, serializeSegments)

	// Validate that when Stop is called, jobs will stop
	assert.Equal(t, result.Cache, "")
	time.AfterFunc(3*time.Second, func() { result.Stop() })
	result.jobs()
	assert.Equal(t, result.Cache, "data from splitChanges and segmentChanges")
}
