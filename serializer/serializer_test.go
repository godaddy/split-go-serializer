package serializer

import (
	"testing"

	"github.com/godaddy/split-go-serializer/api"
	"github.com/stretchr/testify/assert"
)

const (
	testKey           = "someKey"
	serializeSegments = true
)

func TestNewSerializerValid(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	// Act
	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)

	// Validate that returned Serializer has the correct type and values
	assert.Equal(t, result.splitioAPIKey, testKey)
	assert.IsType(t, result.splitioAPIBinding, api.SplitioAPIBinding{})
	assert.Equal(t, result.pollingRateSeconds, pollingRateSeconds)
	assert.Equal(t, result.serializeSegments, serializeSegments)
}

func TestNewSerializerDefaultPollingRateSeconds(t *testing.T) {
	// Arrange
	pollingRateSeconds := 0

	// Act
	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)
	expectedPollingRateSeconds := 300

	// Validate that returned Serializer has the correct default polling rate seconds
	assert.Equal(t, result.pollingRateSeconds, expectedPollingRateSeconds)
}

func TestGetSerializedDataReturnsError(t *testing.T) {
	// Arrange
	pollingRateSeconds := 400

	// Act
	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)
	err := result.GetSerializedData()

	// Validate that GetSerializedData function returns the error
	assert.EqualError(t, err, "not implemented")
}
