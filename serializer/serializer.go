package serializer

import (
	"encoding/json"
	"fmt"

	"github.com/godaddy/split-go-serializer/poller"
)

const formattedLoggingScript = `
<script>
  window.__splitCachePreload = {
	splitsData: %v,
	since: %v,
	segmentsData: %v,
	usingSegmentsCount: %v
  };
</script>`

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
	latestData := serializer.poller.GetCache()

	// splits
	marshalledSplits, err := json.Marshal(latestData.Splits)
	if err != nil {
		return "", err
	}
	stringifiedSplits := string(marshalledSplits)

	// segments
	marshalledSegments, err := json.Marshal(latestData.Segments)
	if err != nil {
		return "", err
	}
	stringifiedSegments := string(marshalledSegments)

	return fmt.Sprintf(formattedLoggingScript, stringifiedSplits, latestData.Since,
		stringifiedSegments, latestData.UsingSegmentsCount), nil
}
