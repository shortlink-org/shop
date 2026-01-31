package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	logger "github.com/shortlink-org/go-sdk/logger"
)

// Notification represents a notification message to be sent to clients
type Notification struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// StockDepletedNotification represents a notification when stock is depleted
type StockDepletedNotification struct {
	GoodID  string `json:"good_id"`
	Message string `json:"message"`
}

// Notifier manages websocket connections and sends notifications
type Notifier struct {
	mu          sync.RWMutex
	connections map[uuid.UUID]*websocket.Conn // customer_id -> connection
	log         logger.Logger
	upgrader    websocket.Upgrader
}

// NewNotifier creates a new websocket notifier
func NewNotifier(log logger.Logger) *Notifier {
	return &Notifier{
		connections: make(map[uuid.UUID]*websocket.Conn),
		log:         log,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// RegisterConnection registers a websocket connection for a customer
func (n *Notifier) RegisterConnection(customerId uuid.UUID, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Close existing connection if any
	if oldConn, exists := n.connections[customerId]; exists {
		_ = oldConn.Close()
	}

	n.connections[customerId] = conn
	n.log.Info("WebSocket connection registered", slog.String("customer_id", customerId.String()))

	// Start ping/pong handler
	go n.handleConnection(customerId, conn)
}

// UnregisterConnection removes a websocket connection
func (n *Notifier) UnregisterConnection(customerId uuid.UUID) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if conn, exists := n.connections[customerId]; exists {
		_ = conn.Close()
		delete(n.connections, customerId)
		n.log.Info("WebSocket connection unregistered", slog.String("customer_id", customerId.String()))
	}
}

// NotifyStockDepleted sends a notification to a customer about stock depletion
func (n *Notifier) NotifyStockDepleted(customerId uuid.UUID, goodId uuid.UUID) error {
	n.mu.RLock()
	conn, exists := n.connections[customerId]
	n.mu.RUnlock()

	if !exists {
		n.log.Warn("No websocket connection found for customer", slog.String("customer_id", customerId.String()))
		return nil // Not an error - customer might not be connected
	}

	notification := Notification{
		Type:    "stock_depleted",
		Message: "Товар закончился на складе и был удален из вашей корзины",
		Data: StockDepletedNotification{
			GoodID:  goodId.String(),
			Message: "Товар закончился на складе и был удален из вашей корзины",
		},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		n.log.Warn("Failed to send notification",
			slog.String("customer_id", customerId.String()),
			slog.String("error", err.Error()))
		// Remove connection on error
		delete(n.connections, customerId)
		return err
	}

	n.log.Info("Stock depletion notification sent",
		slog.String("customer_id", customerId.String()),
		slog.String("good_id", goodId.String()))

	return nil
}

// Upgrade upgrades HTTP connection to websocket
func (n *Notifier) Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return n.upgrader.Upgrade(w, r, nil)
}

// handleConnection handles ping/pong and connection cleanup
func (n *Notifier) handleConnection(customerId uuid.UUID, conn *websocket.Conn) {
	defer n.UnregisterConnection(customerId)

	for {
		messageType, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				n.log.Warn("WebSocket error", slog.String("error", err.Error()))
			}
			break
		}

		// Handle ping/pong
		if messageType == websocket.PingMessage {
			if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
				break
			}
		}
	}
}
