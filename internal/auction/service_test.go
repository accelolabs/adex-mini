package auction

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMatches(t *testing.T) {
	basePartner := Partner{
		IsEnabled:         true,
		Countries:         []string{"RU"},
		DeviceTypes:       []string{"mobile"},
		MinBidFloor:       1,
		BlockedCategories: []string{"gambling"},
	}
	baseAuction := Auction{
		Country:    "RU",
		DeviceType: "mobile",
		BidFloor:   1,
		Categories: []string{"news"},
	}

	tests := []struct {
		name    string
		partner Partner
		auction Auction
		want    bool
	}{
		{name: "all rules pass", partner: basePartner, auction: baseAuction, want: true},
		{name: "disabled", partner: withPartner(basePartner, func(p *Partner) { p.IsEnabled = false }), auction: baseAuction},
		{name: "country rejected", partner: basePartner, auction: withAuction(baseAuction, func(a *Auction) { a.Country = "US" })},
		{name: "empty countries allow any", partner: withPartner(basePartner, func(p *Partner) { p.Countries = nil }), auction: withAuction(baseAuction, func(a *Auction) { a.Country = "US" }), want: true},
		{name: "device rejected", partner: basePartner, auction: withAuction(baseAuction, func(a *Auction) { a.DeviceType = "tv" })},
		{name: "empty devices allow any", partner: withPartner(basePartner, func(p *Partner) { p.DeviceTypes = nil }), auction: withAuction(baseAuction, func(a *Auction) { a.DeviceType = "tv" }), want: true},
		{name: "floor rejected", partner: basePartner, auction: withAuction(baseAuction, func(a *Auction) { a.BidFloor = 0.99 })},
		{name: "equal floor allowed", partner: basePartner, auction: baseAuction, want: true},
		{name: "category blocked", partner: basePartner, auction: withAuction(baseAuction, func(a *Auction) { a.Categories = []string{"news", "gambling"} })},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matches(tt.partner, tt.auction); got != tt.want {
				t.Fatalf("matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceRun(t *testing.T) {
	repository := fakeRepository{partners: []Partner{
		{Code: "success", Endpoint: "success", IsEnabled: true},
		{Code: "error", Endpoint: "error", IsEnabled: true},
		{Code: "filtered", Endpoint: "success", IsEnabled: false},
	}}
	client := fakeDSPClient{}
	service := NewService(repository, client, 50*time.Millisecond)

	result, err := service.Run(context.Background(), Auction{RequestID: "request-1"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Sent != 2 || result.Succeeded != 1 || result.Filtered != 1 {
		t.Fatalf("Run() counts = sent:%d succeeded:%d filtered:%d", result.Sent, result.Succeeded, result.Filtered)
	}
}

func TestServiceTimeout(t *testing.T) {
	repository := fakeRepository{partners: []Partner{{Code: "slow", Endpoint: "timeout", IsEnabled: true}}}
	service := NewService(repository, fakeDSPClient{}, 10*time.Millisecond)

	startedAt := time.Now()
	result, err := service.Run(context.Background(), Auction{RequestID: "request-1"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Succeeded != 0 {
		t.Fatalf("Run() succeeded = %d, want 0", result.Succeeded)
	}
	if elapsed := time.Since(startedAt); elapsed < 10*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Fatalf("Run() elapsed = %s", elapsed)
	}
}

type fakeRepository struct {
	partners []Partner
}

func (r fakeRepository) Partners(context.Context) ([]Partner, error) {
	return r.partners, nil
}

type fakeDSPClient struct{}

func (fakeDSPClient) Send(ctx context.Context, partner Partner, _ Auction) error {
	switch partner.Endpoint {
	case "success":
		return nil
	case "error":
		return errors.New("DSP error")
	case "timeout":
		<-ctx.Done()
		return ctx.Err()
	default:
		return nil
	}
}

func withPartner(partner Partner, change func(*Partner)) Partner {
	change(&partner)
	return partner
}

func withAuction(auction Auction, change func(*Auction)) Auction {
	change(&auction)
	return auction
}
