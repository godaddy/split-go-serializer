package serializer

import (
	"fmt"

	"github.com/godaddy/split-go-serializer/poller"
)

// Serializer contains splitioAPIBinding
type Serializer struct {
	splitioAPIKey string
	poller        poller.Poller
}

// NewSerializer returns a new Serializer
func NewSerializer(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool) *Serializer {
	poller := poller.NewPoller(splitioAPIKey, pollingRateSeconds, serializeSegments)

	return &Serializer{splitioAPIKey, *poller}
}

// GetSerializedData will serialize split and segment data into strings
func (serializer *Serializer) GetSerializedData() error {
	return fmt.Errorf("not implemented")
}
