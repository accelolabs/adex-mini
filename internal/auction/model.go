package auction

import "time"

type Auction struct {
	RequestID  string
	Country    string
	DeviceType string
	BidFloor   float64
	Categories []string
}

type Partner struct {
	UUID              string
	Code              string
	Name              string
	Endpoint          string
	IsEnabled         bool
	Countries         []string
	DeviceTypes       []string
	MinBidFloor       float64
	BlockedCategories []string
}

type Result struct {
	RequestID   string
	MatchedDSPs []string
	Filtered    int
	Sent        int
	Succeeded   int
	Duration    time.Duration
}
