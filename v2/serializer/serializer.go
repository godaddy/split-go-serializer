package serializer

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/godaddy/split-go-serializer/v2/poller"
)

const emptyCacheLoggingScript = `<script>window.__splitCachePreload = {}</script>`

const formattedLoggingScript = `<script>window.__splitCachePreload = { splitsData: %v, since: %v, segmentsData: %v, usingSegmentsCount: %v }</script>`

// SplitCachePreload does something
type SplitCachePreload struct {
	Since              int64
	UsingSegmentsCount int
	SplitsData         string
	SegmentsData       string
}

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
		return emptyCacheLoggingScript, nil
	}
	splitsData := map[string]string{}

	// Serialize values for splits
	for _, split := range latestData.Splits {
		marshalledSplit, _ := json.Marshal(split)
		splitsData[split.Name] = string(marshalledSplit)
	}

	marshalledSplits, err := json.Marshal(splitsData)
	if err != nil {
		return "", err
	}

	segmentsData := map[string]string{}

	// Serialize values for segments
	for _, segment := range latestData.Segments {
		marshalledSegment, _ := json.Marshal(segment)
		segmentsData[segment.Name] = string(marshalledSegment)
	}

	marshalledSegments, err := json.Marshal(segmentsData)
	if err != nil {
		return "", err
	}

	splitCachePreload := &SplitCachePreload{latestData.Since, latestData.UsingSegmentsCount, string(marshalledSplits), string(marshalledSegments)}

	return fmt.Sprintf(formattedLoggingScript, splitCachePreload.SplitsData, splitCachePreload.Since, splitCachePreload.SegmentsData, splitCachePreload.UsingSegmentsCount), nil
}
