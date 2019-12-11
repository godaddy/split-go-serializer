package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testKey = "someKey"

func TestNewSplitioAPIBindingValid(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(testKey)

	// Validate that returned NewSplitioAPIBinding has the correct api key
	assert.EqualValues(t, testKey, result.splitioAPIKey)
}

func TestHTTPGetReturnsError(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(testKey)
	err := result.HTTPGet()

	// Validate that HTTPGet function returns error
	assert.EqualError(t, err, "not implemented")
}

func TestGetSegmentChangesReturnsError(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(testKey)
	err := result.GetSegmentChanges()

	// Validate that GetSegmentChanges function returns error
	assert.EqualError(t, err, "not implemented")
}

func TestGetSplitChangesReturnsError(t *testing.T) {
	// Act
	result := NewSplitioAPIBinding(testKey)
	err := result.GetSplitChanges()

	// Validate that GetSplitChanges function returns error
	assert.EqualError(t, err, "not implemented")
}
