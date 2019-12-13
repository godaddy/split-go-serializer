package serializer

import (
	"fmt"

	"github.com/godaddy/split-go-serializer/api"
)

// Serializer contains splitioAPIBinding
type Serializer struct {
	splitioAPIKey      string
	splitioAPIBinding  api.SplitioAPIBinding
	pollingRateSeconds int
	serializeSegments  bool
}

// NewSerializer returns a new Serializer
func NewSerializer(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool) *Serializer {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	splitioAPIBinding := api.NewSplitioAPIBinding(splitioAPIKey, "")

	return &Serializer{splitioAPIKey, *splitioAPIBinding, pollingRateSeconds, serializeSegments}
}

// GetSerializedData will serialize split and segment data into strings
func (serializer *Serializer) GetSerializedData() error {
	return fmt.Errorf("not implemented")
}
