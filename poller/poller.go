package poller

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/godaddy/split-go-serializer/v2/api"
	"github.com/splitio/go-client/splitio/service/dtos"
)

const emptyCacheLoggingScript = `<script>window.__splitCachePreload = {}</script>`

const formattedLoggingScript = `<script>window.__splitCachePreload = { splitsData: %v, since: %v, segmentsData: %v, usingSegmentsCount: %v }</script>`

// Fetcher is an interface contains GetSplitData, Start and Stop functions
type Fetcher interface {
	Start()
	Stop()
	GetSplitData() SplitData
	GetSerializedData() string
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

// SplitData contains Splits and Segments which is supposed to be updated periodically
type SplitData struct {
	Splits             map[string]dtos.SplitDTO
	Since              int64
	Segments           map[string]dtos.SegmentChangesDTO
	UsingSegmentsCount int
}

// Cache contains raw split data as well as the data in serialized format
type Cache struct {
	SplitData
	SerializedData string
}

// SplitCachePreload does something
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
		SplitData:      SplitData{},
		SerializedData: emptyCacheLoggingScript,
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
	serializedData, err := generateSerializedData(splitData)
	if err != nil {
		poller.Error <- err
		return
	}
	updatedCache := Cache{
		SplitData:      splitData,
		SerializedData: serializedData,
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))

}

// GetSplitData returns split data cache results
func (poller *Poller) GetSplitData() SplitData {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).SplitData
}

// GetSerializedData returns split data cache results
func (poller *Poller) GetSerializedData() string {
	return (*(*Cache)(atomic.LoadPointer(&poller.cache))).SerializedData
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

func generateSerializedData(splitData SplitData) (string, error) {
	if reflect.DeepEqual(splitData, SplitData{}) {
		return emptyCacheLoggingScript, nil
	}
	splitsData := map[string]string{}

	// Serialize values for splits
	for _, split := range splitData.Splits {
		marshalledSplit, _ := json.Marshal(split)
		splitsData[split.Name] = string(marshalledSplit)
	}

	marshalledSplits, err := json.Marshal(splitsData)
	if err != nil {
		return "", err
	}

	segmentsData := map[string]string{}

	// Serialize values for segments
	for _, segment := range splitData.Segments {
		marshalledSegment, _ := json.Marshal(segment)
		segmentsData[segment.Name] = string(marshalledSegment)
	}

	marshalledSegments, err := json.Marshal(segmentsData)
	if err != nil {
		return "", err
	}

	splitCachePreload := &SplitCachePreload{splitData.Since, splitData.UsingSegmentsCount, string(marshalledSplits), string(marshalledSegments)}

	return fmt.Sprintf(formattedLoggingScript, splitCachePreload.SplitsData, splitCachePreload.Since, splitCachePreload.SegmentsData, splitCachePreload.UsingSegmentsCount), nil
}
