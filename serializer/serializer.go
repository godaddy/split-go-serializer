package serializer

import (
	"encoding/json"
	"fmt"

	"github.com/godaddy/split-go-serializer/poller"
)

const formattedLoggingScript = `<script>window.__splitCachePreload = %s</script>`

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
	latestData := serializer.poller.GetSplitData()
	splitCachePreload, err := json.Marshal(latestData)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(formattedLoggingScript, string(splitCachePreload)), nil
}
