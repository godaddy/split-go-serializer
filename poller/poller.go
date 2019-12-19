package poller

import (
	"time"

	"github.com/godaddy/split-go-serializer/api"
)

// Poller contains cache data and SplitioAPIBinding
type Poller struct {
	Cache              int
	splitioAPIBinding  api.SplitioAPIBinding
	pollingRateSeconds int
	serializeSegments  bool
	quit               chan bool
	errorChannel       chan error
	Error              error
}

// NewPoller returns a new Poller
func NewPoller(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool) *Poller {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	splitioAPIBinding := api.NewSplitioAPIBinding(splitioAPIKey, "")
	return &Poller{0, *splitioAPIBinding, pollingRateSeconds, serializeSegments, make(chan bool), make(chan error), nil}
}

// pollForChanges will get the latest data of splits and segment
func (poller *Poller) pollForChanges() {
	// TODO: call getSplits and getSegments to formulate the cach
	// if any of the returned splits/segments have error:
	// 1. pass the error to poller.Error and log the error
	// 2. send the error to poller.errorChannel so it will stop the loop
	poller.Cache++
}

// Start creates a goroutine and keep tracking until it stops
func (poller *Poller) Start() {
	poller.pollForChanges()
	go poller.jobs()
}

// Stop sets quit to true in order to stop the loop
func (poller *Poller) Stop() {
	poller.quit <- true
	poller.quit = make(chan bool)
	poller.errorChannel = make(chan error)
}

// jobs controls whether keep or stop running
func (poller *Poller) jobs() {
	ticker := time.NewTicker(time.Duration(poller.pollingRateSeconds) * time.Second)
	for {
		select {
		case <-poller.quit:
			ticker.Stop()
			return
		case <-poller.errorChannel:
			ticker.Stop()
			return
		case <-ticker.C:
			poller.pollForChanges()
		}
	}
}
