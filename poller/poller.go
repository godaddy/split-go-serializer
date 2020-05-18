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

	"github.com/godaddy/split-go-serializer/v3/api"
	"github.com/splitio/go-client/splitio/service/dtos"
)

const emptyCacheLoggingScript = `<script>window.__splitCachePreload = {}</script>`

const formattedLoggingScript = `<script>window.__splitCachePreload = { splitsData: %v, since: %v, segmentsData: %v, usingSegmentsCount: %v }</script>`

// Fetcher is an interface contains GetSerializedData, Start and Stop functions
type Fetcher interface {
	Start()
	Stop()
	GetSerializedData(splitNames []string) string
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
	splitData             SplitData
	serializedData        string
	serializedDataSubsets map[string]string // key will be a period-delimited string of sorted split names (AKA a subset)
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
		splitData:             SplitData{},
		serializedData:        emptyCacheLoggingScript,
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
	serializedData := poller.generateSerializedData(splitData, []string{})

	updatedCache := Cache{
		splitData:             splitData,
		serializedData:        serializedData,
		serializedDataSubsets: poller.getUpdatedSerializedDataSubsets(splitData),
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))
}

// GetSerializedData returns serialized data cache results
func (poller *Poller) GetSerializedData(splitNames []string) string {
	if len(splitNames) > 0 {
		return poller.getSerializedDataSubset(splitNames)
	}
	return poller.getSerializedData()
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

// getSerializedDataSubset returns serialized data for the splitNames provided
func (poller *Poller) getSerializedDataSubset(splitNames []string) string {
	currentSplitData := poller.getSplitData()
	updatedSubsets := poller.getCachedSerializedDataSubsets()
	sort.Strings(splitNames)
	key := strings.Join(splitNames, ".")

	subset, inMap := updatedSubsets[key]
	if inMap {
		return subset
	}
	subset = poller.generateSerializedData(currentSplitData, splitNames)
	updatedSubsets[key] = subset

	// update cache
	updatedCache := Cache{
		splitData:             currentSplitData,
		serializedData:        poller.getSerializedData(),
		serializedDataSubsets: updatedSubsets,
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))

	return subset
}

// getUpdatedSerializedDataSubsets updates cached serializedDataSubsets based on new split data
func (poller *Poller) getUpdatedSerializedDataSubsets(newSplitData SplitData) map[string]string {
	updatedSubsets := poller.getCachedSerializedDataSubsets()
	for key := range updatedSubsets {
		updatedSubsets[key] = poller.generateSerializedData(newSplitData, strings.Split(key, "."))
	}
	return updatedSubsets
}

// generateSerializedData takes SplitData and generates a script tag
// that saves the SplitData info to the window object of the browser
func (poller *Poller) generateSerializedData(splitData SplitData, splitNames []string) string {
	if reflect.DeepEqual(splitData, SplitData{}) {
		return emptyCacheLoggingScript
	}
	splitNamesToSerializedData := map[string]string{}
	splitsSubset := map[string]dtos.SplitDTO{}
	serializingASubsetOfSplits := len(splitNames) > 0

	// Serialize values for splits
	for name, split := range splitData.Splits {
		if serializingASubsetOfSplits {
			index := sort.SearchStrings(splitNames, split.Name)
			splitIsInSplitNames := index < len(splitNames) && splitNames[index] == split.Name
			// if the split is not in the splitNames array, do not serialize the split
			if !splitIsInSplitNames {
				continue
			}
			splitsSubset[name] = split
		}
		marshalledSplit, _ := json.Marshal(split)
		splitNamesToSerializedData[split.Name] = string(marshalledSplit)
	}

	marshalledSplits, _ := json.Marshal(splitNamesToSerializedData)

	segments := splitData.Segments
	usingSegmentsCount := splitData.UsingSegmentsCount

	// get segments and usingSegmentsCount for subset of splits
	if poller.serializeSegments && serializingASubsetOfSplits {
		var err error
		binding := poller.splitio
		segments, usingSegmentsCount, err = binding.GetSegmentsForSplits(splitsSubset)
		if err != nil {
			return emptyCacheLoggingScript
		}
	}

	segmentsData := map[string]string{}

	// Serialize values for segments
	for _, segment := range segments {
		marshalledSegment, _ := json.Marshal(segment)
		segmentsData[segment.Name] = string(marshalledSegment)
	}

	marshalledSegments, _ := json.Marshal(segmentsData)

	splitCachePreload := &SplitCachePreload{splitData.Since, usingSegmentsCount, string(marshalledSplits), string(marshalledSegments)}

	return fmt.Sprintf(formattedLoggingScript, splitCachePreload.SplitsData, splitCachePreload.Since, splitCachePreload.SegmentsData, splitCachePreload.UsingSegmentsCount)
}

// getSplitData returns cached split data
func (poller *Poller) getSplitData() SplitData {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).splitData
}

// getSerializedData returns cached serialized data
func (poller *Poller) getSerializedData() string {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).serializedData
}

// getCachedSerializedDataSubsets returns cached serialized data for split subsets
func (poller *Poller) getCachedSerializedDataSubsets() map[string]string {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).serializedDataSubsets
}
