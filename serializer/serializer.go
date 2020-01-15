package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"

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
	if reflect.DeepEqual(latestData, poller.SplitData{}) {
		return fmt.Sprintf(formattedLoggingScript, "{}"), nil
	}

	splitCachePreload := map[string]interface{}{"since": latestData.Since, "usingSegmentsCount": latestData.UsingSegmentsCount, "splitsData": map[string]string{}, "segmentsData": map[string]string{}}

	// Serialize values for splits
	for _, split := range latestData.Splits {
		marshalledSplit, err := json.Marshal(split)
		if err != nil {
			return "", err
		}
		splitCachePreload["splitsData"].(map[string]string)[split.Name] = string(marshalledSplit)
	}

	// Serialize values for segments
	for _, segment := range latestData.Segments {
		marshalledSegment, err := json.Marshal(segment)
		if err != nil {
			return "", err
		}
		splitCachePreload["segmentsData"].(map[string]string)[segment.Name] = string(marshalledSegment)
	}

	marshalledSplitCachePreload, err := json.Marshal(splitCachePreload)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(formattedLoggingScript, string(marshalledSplitCachePreload)), nil
}
