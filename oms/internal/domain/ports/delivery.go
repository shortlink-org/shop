package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DeliveryClient defines the interface for communicating with the Delivery service.
// This is a port in the hexagonal architecture that will be implemented
// by a gRPC adapter in the infrastructure layer.
type DeliveryClient interface {
	// AcceptOrder sends an order to the Delivery service for processing.
	// Returns the package ID assigned by the Delivery service.
	AcceptOrder(ctx context.Context, req AcceptOrderRequest) (*AcceptOrderResponse, error)
}

// AcceptOrderRequest contains the data needed to request delivery.
type AcceptOrderRequest struct {
	// OrderID is the unique identifier of the order
	OrderID uuid.UUID
	// CustomerID is the unique identifier of the customer
	CustomerID uuid.UUID
	// PickupAddress is where the package should be picked up
	PickupAddress DeliveryAddress
	// DeliveryAddress is where the package should be delivered
	DeliveryAddress DeliveryAddress
	// DeliveryPeriod is the desired delivery time window
	DeliveryPeriod DeliveryPeriodDTO
	// PackageInfo contains physical characteristics of the package
	PackageInfo PackageInfoDTO
	// Priority indicates the delivery priority level
	Priority DeliveryPriorityDTO
}

// AcceptOrderResponse contains the response from the Delivery service.
type AcceptOrderResponse struct {
	// PackageID is the unique identifier assigned by the Delivery service
	PackageID string
	// Status is the initial package status
	Status string
}

// DeliveryAddress represents a physical address for delivery.
type DeliveryAddress struct {
	Street     string
	City       string
	PostalCode string
	Country    string
	Latitude   float64
	Longitude  float64
}

// DeliveryPeriodDTO represents the desired delivery time window.
type DeliveryPeriodDTO struct {
	StartTime time.Time
	EndTime   time.Time
}

// PackageInfoDTO contains physical characteristics of the package.
type PackageInfoDTO struct {
	WeightKg   float64
	Dimensions string
}

// DeliveryPriorityDTO represents delivery priority level.
type DeliveryPriorityDTO int32

const (
	DeliveryPriorityUnspecified DeliveryPriorityDTO = 0
	DeliveryPriorityNormal      DeliveryPriorityDTO = 1
	DeliveryPriorityUrgent      DeliveryPriorityDTO = 2
)
