package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testKey = "someKey"

func TestNewSplitioAPIBindingValid(t *testing.T) {
	result := NewSplitioAPIBinding(testKey)
	assert.EqualValues(t, testKey, result.splitioAPIKey)
}

func TestHTTPGetReturnsError(t *testing.T) {

	result := NewSplitioAPIBinding(testKey)
	err := result.HTTPGet()

	assert.EqualError(t, err, "not implemented")
}

func TestGetSegmentChangesReturnsError(t *testing.T) {

	result := NewSplitioAPIBinding(testKey)
	err := result.GetSegmentChanges()

	assert.EqualError(t, err, "not implemented")
}

func TestGetSplitDataReturnsError(t *testing.T) {

	result := NewSplitioAPIBinding(testKey)
	err := result.GetSplitData()

	assert.EqualError(t, err, "not implemented")
}
