package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	logger "github.com/shortlink-org/go-sdk/logger"

	on_stock_changed "github.com/shortlink-org/shop/oms/internal/usecases/cart/event/on_stock_changed"
)

// StockChangeHandler handles HTTP requests for stock change events
type StockChangeHandler struct {
	log                   logger.Logger
	stockChangedHandler   *on_stock_changed.Handler
}

// NewStockChangeHandler creates a new stock change event handler
func NewStockChangeHandler(log logger.Logger, stockChangedHandler *on_stock_changed.Handler) *StockChangeHandler {
	return &StockChangeHandler{
		log:                   log,
		stockChangedHandler:   stockChangedHandler,
	}
}

// StockChangeRequest represents the incoming stock change event
type StockChangeRequest struct {
	GoodID      string `json:"good_id"`
	OldQuantity uint32 `json:"old_quantity"`
	NewQuantity uint32 `json:"new_quantity"`
}

// Handle handles stock change events
func (h *StockChangeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StockChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("Failed to decode stock change request", slog.String("error", err.Error()))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	goodId, err := uuid.Parse(req.GoodID)
	if err != nil {
		h.log.Warn("Invalid good ID in stock change request", slog.String("good_id", req.GoodID), slog.String("error", err.Error()))
		http.Error(w, "Invalid good_id", http.StatusBadRequest)
		return
	}

	// Handle stock change - remove item from carts if stock is zero
	ctx := r.Context()
	event := on_stock_changed.NewEvent(goodId, req.NewQuantity)
	if err := h.stockChangedHandler.Handle(ctx, event); err != nil {
		h.log.Error("Failed to handle stock change", slog.String("error", err.Error()))
		// Don't return error to caller - event was received, processing may continue
		// In production, you might want to return 202 Accepted and process asynchronously
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
