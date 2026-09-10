package auction

import (
	"errors"
	"strings"
)

type auctionRequest struct {
	RequestID  string   `json:"request_id"`
	Country    string   `json:"country"`
	DeviceType string   `json:"device_type"`
	BidFloor   float64  `json:"bid_floor"`
	Categories []string `json:"categories"`
}

type auctionResponse struct {
	RequestID   string   `json:"request_id"`
	MatchedDSPs []string `json:"matched_dsps"`
	Sent        int      `json:"sent"`
	Succeeded   int      `json:"succeeded"`
	DurationMS  int64    `json:"duration_ms"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (r auctionRequest) validate() error {
	if strings.TrimSpace(r.RequestID) == "" {
		return errors.New("request_id is required")
	}
	if len(r.Country) != 2 || r.Country[0] < 'A' || r.Country[0] > 'Z' || r.Country[1] < 'A' || r.Country[1] > 'Z' {
		return errors.New("country must contain two uppercase letters")
	}
	switch r.DeviceType {
	case "mobile", "desktop", "tv":
	default:
		return errors.New("device_type must be mobile, desktop, or tv")
	}
	if r.BidFloor < 0 {
		return errors.New("bid_floor must be greater than or equal to zero")
	}
	return nil
}

func (r auctionRequest) toModel() Auction {
	return Auction{
		RequestID:  r.RequestID,
		Country:    r.Country,
		DeviceType: r.DeviceType,
		BidFloor:   r.BidFloor,
		Categories: r.Categories,
	}
}
