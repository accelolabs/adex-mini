package auction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const maxRequestBodySize = 1 << 20

type AuctionService interface {
	Run(ctx context.Context, auction Auction) (Result, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	service AuctionService
	health  HealthChecker
	logger  *slog.Logger
}

func NewHandler(service AuctionService, health HealthChecker, logger *slog.Logger) *Handler {
	return &Handler{service: service, health: health, logger: logger}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auction", h.handleAuction)
	mux.HandleFunc("GET /healthz", h.handleHealth)
	return mux
}

func (h *Handler) handleAuction(w http.ResponseWriter, r *http.Request) {
	startedAt := time.Now()
	request, err := decodeAuctionRequest(w, r)
	if err != nil {
		h.logger.Warn("invalid auction request", "request_id", request.RequestID, "error", err, "duration_ms", time.Since(startedAt).Milliseconds())
		h.writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.Run(r.Context(), request.toModel())
	if err != nil {
		h.logger.Error("auction failed", "request_id", request.RequestID, "error", err, "duration_ms", time.Since(startedAt).Milliseconds())
		h.writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	h.logger.Info(
		"auction completed",
		"request_id", result.RequestID,
		"matched", len(result.MatchedDSPs),
		"filtered", result.Filtered,
		"sent", result.Sent,
		"succeeded", result.Succeeded,
		"duration_ms", result.Duration.Milliseconds(),
	)

	h.writeJSON(w, http.StatusOK, auctionResponse{
		RequestID:   result.RequestID,
		MatchedDSPs: result.MatchedDSPs,
		Sent:        result.Sent,
		Succeeded:   result.Succeeded,
		DurationMS:  result.Duration.Milliseconds(),
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := h.health.Ping(ctx); err != nil {
		h.logger.Error("health check failed", "error", err)
		h.writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "service unavailable"})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeAuctionRequest(w http.ResponseWriter, r *http.Request) (auctionRequest, error) {
	var request auctionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodySize))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		return request, fmt.Errorf("decode JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return request, errors.New("request body must contain one JSON object")
	}
	if err := request.validate(); err != nil {
		return request, err
	}

	return request, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("write JSON response", "error", err)
	}
}
