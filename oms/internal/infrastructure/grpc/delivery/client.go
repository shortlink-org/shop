package delivery

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

// Client implements the ports.DeliveryClient interface using gRPC.
type Client struct {
	conn   *grpc.ClientConn
	client DeliveryServiceClient
}

// Config contains configuration for the Delivery gRPC client.
type Config struct {
	// Address is the gRPC server address (e.g., "delivery:50051")
	Address string
	// Timeout is the connection timeout
	Timeout time.Duration
}

// NewClient creates a new Delivery gRPC client.
func NewClient(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to delivery service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: NewDeliveryServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// AcceptOrder sends an order to the Delivery service for processing.
func (c *Client) AcceptOrder(ctx context.Context, req ports.AcceptOrderRequest) (*ports.AcceptOrderResponse, error) {
	grpcReq := &AcceptOrderRequest{
		OrderId:    req.OrderID.String(),
		CustomerId: req.CustomerID.String(),
		PickupAddress: &Address{
			Street:     req.PickupAddress.Street,
			City:       req.PickupAddress.City,
			PostalCode: req.PickupAddress.PostalCode,
			Country:    req.PickupAddress.Country,
			Latitude:   req.PickupAddress.Latitude,
			Longitude:  req.PickupAddress.Longitude,
		},
		DeliveryAddress: &Address{
			Street:     req.DeliveryAddress.Street,
			City:       req.DeliveryAddress.City,
			PostalCode: req.DeliveryAddress.PostalCode,
			Country:    req.DeliveryAddress.Country,
			Latitude:   req.DeliveryAddress.Latitude,
			Longitude:  req.DeliveryAddress.Longitude,
		},
		DeliveryPeriod: &DeliveryPeriod{
			StartTime: timestamppb.New(req.DeliveryPeriod.StartTime),
			EndTime:   timestamppb.New(req.DeliveryPeriod.EndTime),
		},
		PackageInfo: &PackageInfo{
			WeightKg: req.PackageInfo.WeightKg,
		},
		Priority: Priority(req.Priority),
		RecipientContacts: &RecipientContacts{
			RecipientName:  req.RecipientName,
			RecipientPhone: req.RecipientPhone,
			RecipientEmail: req.RecipientEmail,
		},
	}

	resp, err := c.client.AcceptOrder(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to accept order: %w", err)
	}

	return &ports.AcceptOrderResponse{
		PackageID: resp.PackageId,
		Status:    packageStatusToString(resp.Status),
	}, nil
}

// packageStatusToString converts proto PackageStatus to string.
func packageStatusToString(status PackageStatus) string {
	switch status {
	case PackageStatus_PACKAGE_STATUS_ACCEPTED:
		return "ACCEPTED"
	case PackageStatus_PACKAGE_STATUS_IN_POOL:
		return "IN_POOL"
	case PackageStatus_PACKAGE_STATUS_ASSIGNED:
		return "ASSIGNED"
	case PackageStatus_PACKAGE_STATUS_IN_TRANSIT:
		return "IN_TRANSIT"
	case PackageStatus_PACKAGE_STATUS_DELIVERED:
		return "DELIVERED"
	case PackageStatus_PACKAGE_STATUS_NOT_DELIVERED:
		return "NOT_DELIVERED"
	case PackageStatus_PACKAGE_STATUS_REQUIRES_HANDLING:
		return "REQUIRES_HANDLING"
	default:
		return "UNSPECIFIED"
	}
}
