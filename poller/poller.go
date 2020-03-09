package poller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/godaddy/split-go-serializer/v2/api"
	"github.com/splitio/go-client/splitio/service/dtos"
)

const emptyCacheLoggingScript = `<script>window.__splitCachePreload = {}</script>`

const formattedLoggingScript = `<script>window.__splitCachePreload = { splitsData: %v, since: %v, segmentsData: %v, usingSegmentsCount: %v }</script>`

// Fetcher is an interface contains GetSerializedData, Start and Stop functions
type Fetcher interface {
	Start()
	Stop()
	GetSerializedData() string
	GetSerializedDataSubset([]string) string
}

// Poller implements Fetcher and contains cache pointer, splitio, and required info to interact with aplitio api
type Poller struct {
	Error              chan error
	splitio            api.Splitio
	pollingRateSeconds int
	serializeSegments  bool
	quit               chan bool
	cache              unsafe.Pointer
}

// Cache contains raw split data as well as the data in serialized format
type Cache struct {
	SplitData
	SerializedData        string
	serializedDataSubsets map[string]string
}

// SplitData contains Splits and Segments which is supposed to be updated periodically
type SplitData struct {
	Splits             map[string]dtos.SplitDTO
	Since              int64
	Segments           map[string]dtos.SegmentChangesDTO
	UsingSegmentsCount int
}

// SplitCachePreload contains the same information as SplitData but in string format
type SplitCachePreload struct {
	Since              int64
	UsingSegmentsCount int
	SplitsData         string
	SegmentsData       string
}

// NewPoller returns a new Poller
func NewPoller(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool, splitio api.Splitio) *Poller {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	if splitio == nil {
		splitio = api.NewSplitioAPIBinding(splitioAPIKey, "")
	}
	emptyCache := Cache{
		SplitData:             SplitData{},
		SerializedData:        emptyCacheLoggingScript,
		serializedDataSubsets: make(map[string]string),
	}
	return &Poller{make(chan error), splitio, pollingRateSeconds, serializeSegments, make(chan bool), unsafe.Pointer(&emptyCache)}
}

// pollForChanges updates the Cache with latest splits and segment
func (poller *Poller) pollForChanges() {
	binding := poller.splitio
	splits, since, err := binding.GetSplits()
	if err != nil {
		poller.Error <- err
		return
	}

	segments := map[string]dtos.SegmentChangesDTO{}
	usingSegmentsCount := 0
	if poller.serializeSegments {
		segments, usingSegmentsCount, err = binding.GetSegmentsForSplits(splits)
		if err != nil {
			poller.Error <- err
			return
		}
	}

	// Update Cache
	splitData := SplitData{
		Splits:             splits,
		Since:              since,
		Segments:           segments,
		UsingSegmentsCount: usingSegmentsCount,
	}
	serializedData := generateSerializedData(splitData, []string{})

	updatedCache := Cache{
		SplitData:             splitData,
		SerializedData:        serializedData,
		serializedDataSubsets: poller.getUpdatedSerializedDataSubsets(splitData),
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))

}

// GetSerializedData returns serialized data cache results
func (poller *Poller) GetSerializedData() string {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).SerializedData
}

// GetSerializedDataSubset returns serialized data cache results for the subset of splits
func (poller *Poller) GetSerializedDataSubset(splits []string) string {
	currentSplitData := getSplitData(poller)
	updatedSubsets := poller.getCachedSerializedDataSubsets()
	sort.Strings(splits)
	key := strings.Join(splits, ".")

	subset, inMap := updatedSubsets[key]
	if inMap {
		return subset
	}
	subset = generateSerializedData(currentSplitData, splits)
	updatedSubsets[key] = subset

	// update cache
	updatedCache := Cache{
		SplitData:             currentSplitData,
		SerializedData:        poller.GetSerializedData(),
		serializedDataSubsets: updatedSubsets,
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))

	return subset
}

// Start creates a goroutine and keep tracking until it stops
func (poller *Poller) Start() {
	poller.pollForChanges()
	go poller.jobs()
}

// Stop sets quit to true in order to stop the loop
func (poller *Poller) Stop() {
	poller.quit <- true
}

// jobs controls whether keep or stop running
func (poller *Poller) jobs() {
	ticker := time.NewTicker(time.Duration(poller.pollingRateSeconds) * time.Second)
	for {
		select {
		case <-poller.quit:
			ticker.Stop()
			return
		case <-ticker.C:
			poller.pollForChanges()
		}
	}
}

// getUpdatedSerializedDataSubsets updates cached serializedDataSubsets based on new splits data
func (poller *Poller) getUpdatedSerializedDataSubsets(newSplitData SplitData) map[string]string {
	updatedSubsets := poller.getCachedSerializedDataSubsets()
	for key := range updatedSubsets {
		updatedSubsets[key] = generateSerializedData(newSplitData, strings.Split(key, "."))
	}
	return updatedSubsets
}

// getCachedSerializedDataSubsets returns splits data subsets that are cached
func (poller *Poller) getCachedSerializedDataSubsets() map[string]string {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).serializedDataSubsets
}

// generateSerializedData takes SplitData and generates a script tag
// that saves the SplitData info to the window object of the browser
func generateSerializedData(splitData SplitData, splits []string) string {
	if reflect.DeepEqual(splitData, SplitData{}) {
		return emptyCacheLoggingScript
	}
	splitsData := map[string]string{}

	// Serialize values for splits
	for _, split := range splitData.Splits {
		splitIndex := sort.SearchStrings(splits, split.Name)
		splitInSplits := splitIndex < len(splits) && splits[splitIndex] == split.Name
		// if the split is not in the splits array, do not serialize the split
		if len(splits) > 0 && !splitInSplits {
			continue
		}
		marshalledSplit, _ := json.Marshal(split)
		splitsData[split.Name] = string(marshalledSplit)
	}

	marshalledSplits, _ := json.Marshal(splitsData)

	segmentsData := map[string]string{}

	// Serialize values for segments
	for _, segment := range splitData.Segments {
		marshalledSegment, _ := json.Marshal(segment)
		segmentsData[segment.Name] = string(marshalledSegment)
	}

	marshalledSegments, _ := json.Marshal(segmentsData)

	splitCachePreload := &SplitCachePreload{splitData.Since, splitData.UsingSegmentsCount, string(marshalledSplits), string(marshalledSegments)}

	return fmt.Sprintf(formattedLoggingScript, splitCachePreload.SplitsData, splitCachePreload.Since, splitCachePreload.SegmentsData, splitCachePreload.UsingSegmentsCount)
}

// getSplitData helper returns split data cache results
func getSplitData(poller *Poller) SplitData {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).SplitData
}
