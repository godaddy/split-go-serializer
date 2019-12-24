package poller

import (
	"time"

	"github.com/godaddy/split-go-serializer/api"
	"github.com/splitio/go-client/splitio/service/dtos"
)

// Poller contains cache data, splitioDataGetter, and required info to interact with aplitio api
type Poller struct {
	Cache              Cache
	Error              chan error
	splitioDataGetter  api.SplitioDataGetter
	pollingRateSeconds int
	serializeSegments  bool
	quit               chan bool
}

// Cache contains Splits and Segments which is supposed to be updated periodically
type Cache struct {
	Splits             []dtos.SplitDTO
	Since              int64
	Segments           []dtos.SegmentChangesDTO
	UsingSegmentsCount int
}

// NewPoller returns a new Poller
func NewPoller(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool, splitioDataGetter api.SplitioDataGetter) *Poller {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	if splitioDataGetter != nil {
		return &Poller{Cache{}, make(chan error), splitioDataGetter, pollingRateSeconds, serializeSegments, make(chan bool)}
	}
	splitioAPIBinding := api.NewSplitioAPIBinding(splitioAPIKey, "")
	return &Poller{Cache{}, make(chan error), splitioAPIBinding, pollingRateSeconds, serializeSegments, make(chan bool)}
}

// pollForChanges updates the Cache with latest splits and segment
func (poller *Poller) pollForChanges() {
	binding := poller.splitioDataGetter
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

	poller.Cache = Cache{
		Splits:             splits,
		Since:              since,
		Segments:           segments,
		UsingSegmentsCount: usingSegmentsCount,
	}

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
