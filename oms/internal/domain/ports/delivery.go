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
	// RecipientName is the name of the person receiving the delivery (optional)
	RecipientName string
	// RecipientPhone is the phone number for delivery contact (optional)
	RecipientPhone string
	// RecipientEmail is the email for delivery notifications (optional)
	RecipientEmail string
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
	WeightKg float64
}

// DeliveryPriorityDTO represents delivery priority level.
type DeliveryPriorityDTO int32

const (
	DeliveryPriorityUnspecified DeliveryPriorityDTO = 0
	DeliveryPriorityNormal      DeliveryPriorityDTO = 1
	DeliveryPriorityUrgent      DeliveryPriorityDTO = 2
)
