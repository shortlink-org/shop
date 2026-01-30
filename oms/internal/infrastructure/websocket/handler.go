package websocket

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	logger "github.com/shortlink-org/go-sdk/logger"
)

// Handler handles websocket connections
type Handler struct {
	notifier *Notifier
	log      logger.Logger
}

// NewHandler creates a new websocket handler
func NewHandler(notifier *Notifier, log logger.Logger) *Handler {
	return &Handler{
		notifier: notifier,
		log:      log,
	}
}

// HandleWebSocket handles websocket upgrade requests
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract customer ID from query parameter or header
	customerIdStr := r.URL.Query().Get("customer_id")
	if customerIdStr == "" {
		// Try to get from header
		customerIdStr = r.Header.Get("X-Customer-ID")
	}

	if customerIdStr == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	customerId, err := uuid.Parse(customerIdStr)
	if err != nil {
		http.Error(w, "invalid customer_id", http.StatusBadRequest)
		return
	}

	// Upgrade connection to websocket
	conn, err := h.notifier.Upgrade(w, r)
	if err != nil {
		h.log.Warn("Failed to upgrade connection", slog.String("error", err.Error()))
		return
	}

	// Register connection
	h.notifier.RegisterConnection(customerId, conn)
}

