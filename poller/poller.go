package poller

import (
	"github.com/godaddy/split-go-serializer/api"
	"github.com/jasonlvhit/gocron"
)

// Poller contains cache data and SplitioAPIBinding
type Poller struct {
	Cache              string
	SplitioAPIBinding  api.SplitioAPIBinding
	PollingRateSeconds int
	SerializeSegments  bool
	quit               chan bool
}

// NewPoller returns a new Poller
func NewPoller(splitioAPIKey string, pollingRateSeconds int, serializeSegments bool) *Poller {
	if pollingRateSeconds == 0 {
		pollingRateSeconds = 300
	}
	splitioAPIBinding := api.NewSplitioAPIBinding(splitioAPIKey, "")
	return &Poller{"", *splitioAPIBinding, pollingRateSeconds, serializeSegments, make(chan bool)}
}

// pollForChanges will get the latest data of splits and segment
func (poller *Poller) pollForChanges() error {
	//TODO: call getSplits and getSegments to formulate the cache
	poller.Cache = "data from splitChanges and segmentChanges"
	return nil
}

// Poll creates a goroutine and keep tracking until it stops
func (poller *Poller) Poll() {
	go poller.jobs()
	<-poller.quit
}

// Stop sets quit to true in order to stop the loop
func (poller *Poller) Stop() {
	poller.quit <- true
}

// jobs controls whether keep or stop running
func (poller *Poller) jobs() {
	for {
		g := gocron.NewScheduler()
		g.Every(uint64(poller.PollingRateSeconds)).Seconds().Do(poller.pollForChanges)
		select {
		case <-poller.quit:
			g.Clear()
			close(poller.quit)
			return
		case <-g.Start():
		}
	}
}
