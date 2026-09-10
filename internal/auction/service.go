package auction

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Repository interface {
	Partners(ctx context.Context) ([]Partner, error)
}

type DSPClient interface {
	Send(ctx context.Context, partner Partner, auction Auction) error
}

type Service struct {
	repository Repository
	dspClient  DSPClient
	timeout    time.Duration
}

func NewService(repository Repository, dspClient DSPClient, timeout time.Duration) *Service {
	return &Service{
		repository: repository,
		dspClient:  dspClient,
		timeout:    timeout,
	}
}

func (s *Service) Run(ctx context.Context, request Auction) (Result, error) {
	startedAt := time.Now()
	result := Result{
		RequestID:   request.RequestID,
		MatchedDSPs: make([]string, 0),
	}

	partners, err := s.repository.Partners(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("load partners: %w", err)
	}

	matched := make([]Partner, 0, len(partners))
	for _, partner := range partners {
		if matches(partner, request) {
			matched = append(matched, partner)
			result.MatchedDSPs = append(result.MatchedDSPs, partner.Code)
		}
	}

	result.Filtered = len(partners) - len(matched)
	result.Sent = len(matched)
	if len(matched) == 0 {
		result.Duration = time.Since(startedAt)
		return result, nil
	}

	auctionCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	results := make(chan bool, len(matched))
	var workers sync.WaitGroup
	workers.Add(len(matched))

	for _, partner := range matched {
		go func() {
			defer workers.Done()
			results <- s.dspClient.Send(auctionCtx, partner, request) == nil
		}()
	}

	workers.Wait()
	close(results)

	for succeeded := range results {
		if succeeded {
			result.Succeeded++
		}
	}

	result.Duration = time.Since(startedAt)
	return result, nil
}

func matches(partner Partner, request Auction) bool {
	return partner.IsEnabled &&
		containsOrAny(partner.Countries, request.Country) &&
		containsOrAny(partner.DeviceTypes, request.DeviceType) &&
		request.BidFloor >= partner.MinBidFloor &&
		!intersects(partner.BlockedCategories, request.Categories)
}

func containsOrAny(values []string, target string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func intersects(left, right []string) bool {
	for _, leftValue := range left {
		for _, rightValue := range right {
			if leftValue == rightValue {
				return true
			}
		}
	}
	return false
}

type MockDSPClient struct{}

func NewMockDSPClient() *MockDSPClient {
	return &MockDSPClient{}
}

func (c *MockDSPClient) Send(ctx context.Context, partner Partner, _ Auction) error {
	switch partner.Endpoint {
	case "mock://success":
		return nil
	case "mock://error":
		return errors.New("mock DSP error")
	case "mock://timeout":
		<-ctx.Done()
		return ctx.Err()
	default:
		return fmt.Errorf("unsupported mock endpoint %q", partner.Endpoint)
	}
}
