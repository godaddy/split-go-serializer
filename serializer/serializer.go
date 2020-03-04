package serializer

import (
	"github.com/godaddy/split-go-serializer/v2/poller"
)

// Serializer contains poller
type Serializer struct {
	poller poller.Fetcher
}

// NewSerializer returns a new Serializer
func NewSerializer(poller poller.Fetcher) *Serializer {
	return &Serializer{poller}
}

// GetSerializedData serializes split and segment data into strings
func (serializer *Serializer) GetSerializedData() (string, error) {
	serializedData := serializer.poller.GetSerializedData()
	return serializedData, nil
}
