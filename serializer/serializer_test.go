package serializer

import (
	"testing"

	"github.com/godaddy/split-go-serializer/poller"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

func TestNewSerializerValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400
	testPoller := poller.NewPoller(testKey, pollingRateSeconds, serializeSegments, nil)

	// Act
	result := NewSerializer(testPoller)

	// Validate that returned Serializer has the correct type and values
	assert.IsType(t, result.poller, poller.Poller{})
}

func TestGetSerializedDataReturnsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400
	testPoller := poller.NewPoller(testKey, pollingRateSeconds, serializeSegments, nil)

	// Act
	result := NewSerializer(testPoller)
	err := result.GetSerializedData()

	// Validate that GetSerializedData function returns the error
	assert.EqualError(t, err, "not implemented")
}
