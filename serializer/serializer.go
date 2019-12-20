package serializer

import (
	"fmt"

	"github.com/godaddy/split-go-serializer/poller"
)

// Serializer contains poller
type Serializer struct {
	poller poller.Poller
}

// NewSerializer returns a new Serializer
func NewSerializer(poller *poller.Poller) *Serializer {
	return &Serializer{*poller}
}

// GetSerializedData will serialize split and segment data into strings
func (serializer *Serializer) GetSerializedData() error {
	return fmt.Errorf("not implemented")
}
