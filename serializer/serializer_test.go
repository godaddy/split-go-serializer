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
	pollingRateSeconds := 400

	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)

	assert.Equal(t, result.splitioAPIKey, testKey)
	assert.IsType(t, result.splitioAPIBinding, api.SplitioAPIBinding{})
	assert.Equal(t, result.pollingRateSeconds, pollingRateSeconds)
	assert.Equal(t, result.serializeSegments, serializeSegments)

}

func TestNewSerializerDefaultPollingRateSeconds(t *testing.T) {
	pollingRateSeconds := 0

	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)
	expectedPollingRateSeconds := 300

	assert.Equal(t, result.pollingRateSeconds, expectedPollingRateSeconds)

}

func TestGetSerializedDataReturnsError(t *testing.T) {
	pollingRateSeconds := 400

	result := NewSerializer(testKey, pollingRateSeconds, serializeSegments)
	err := result.GetSerializedData()

	assert.EqualError(t, err, "not implemented")
}
