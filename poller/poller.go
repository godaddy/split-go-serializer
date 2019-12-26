package poller

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/godaddy/split-go-serializer/api"
	"github.com/splitio/go-client/splitio/service/dtos"
)

// Fetcher is an interface contains GetCache, Start and Stop functions
type Fetcher interface {
	Start()
	Stop()
	GetCache() SplitData
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
	Splits             []dtos.SplitDTO
	Since              int64
	Segments           []dtos.SegmentChangesDTO
	UsingSegmentsCount int
}

// NewPoller returns a new Poller
func NewPoller(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool, splitio api.Splitio) *Poller {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	if splitio == nil {
		splitio = api.NewSplitioAPIBinding(splitioAPIKey, "")
	}
	return &Poller{make(chan error), splitio, pollingRateSeconds, serializeSegments, make(chan bool), unsafe.Pointer(&SplitData{})}
}

// pollForChanges updates the Cache with latest splits and segment
func (poller *Poller) pollForChanges() {
	binding := poller.splitio
	splits, since, err := binding.GetSplits()
	if err != nil {
		poller.Error <- err
		return
	}

	segments := []dtos.SegmentChangesDTO{}
	usingSegmentsCount := 0
	if poller.serializeSegments {
		segments, usingSegmentsCount, err = binding.GetSegmentsForSplits(splits)
		if err != nil {
			poller.Error <- err
			return
		}
	}

	// Update Cache
	updatedCache := SplitData{
		Splits:             splits,
		Since:              since,
		Segments:           segments,
		UsingSegmentsCount: usingSegmentsCount,
	}
	atomic.StorePointer(&poller.cache, unsafe.Pointer(&updatedCache))

}

// GetCache returns cache results of SplitData
func (poller *Poller) GetCache() SplitData {
	return *(*SplitData)(atomic.LoadPointer(&poller.cache))
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
