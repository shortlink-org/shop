package pricer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	pricerv1 "github.com/shortlink-org/shop/oms/internal/infrastructure/grpc/pricer/v1"
)

// Client implements the ports.PricerClient interface using gRPC.
type Client struct {
	conn   *grpc.ClientConn
	client pricerv1.CartServiceClient
}

// Config contains configuration for the Pricer gRPC client.
type Config struct {
	// Address is the gRPC server address (e.g., "internal-gateway:443")
	Address string
	// Timeout is the connection timeout
	Timeout time.Duration
	// TLSEnabled enables TLS for the connection
	TLSEnabled bool
	// CertPath is the path to the CA certificate
	CertPath string
}

// NewClient creates a new Pricer gRPC client.
func NewClient(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var opts []grpc.DialOption

	if cfg.TLSEnabled && cfg.CertPath != "" {
		creds, err := loadTLSCredentials(cfg.CertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pricer service: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pricerv1.NewCartServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CalculateTotal calculates the total price, tax, and discounts for a cart.
func (c *Client) CalculateTotal(ctx context.Context, req ports.CalculateTotalRequest) (*ports.CalculateTotalResponse, error) {
	// Convert domain request to proto
	protoReq := &pricerv1.CalculateTotalRequest{
		Cart: &pricerv1.Cart{
			CustomerId: req.Cart.CustomerID.String(),
			Items:      make([]*pricerv1.CartItem, 0, len(req.Cart.Items)),
		},
		DiscountParams: req.DiscountParams,
		TaxParams:      req.TaxParams,
	}

	for _, item := range req.Cart.Items {
		protoReq.Cart.Items = append(protoReq.Cart.Items, &pricerv1.CartItem{
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
			Price:     item.UnitPrice.String(),
		})
	}

	resp, err := c.client.CalculateTotal(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total: %w", err)
	}

	// Convert proto response to domain
	totalTax, _ := decimal.NewFromString(resp.Total.GetTotalTax())
	totalDiscount, _ := decimal.NewFromString(resp.Total.GetTotalDiscount())
	finalPrice, _ := decimal.NewFromString(resp.Total.GetFinalPrice())

	// Calculate subtotal from items
	subtotal := decimal.Zero
	for _, item := range req.Cart.Items {
		subtotal = subtotal.Add(item.UnitPrice.Mul(decimal.NewFromInt32(item.Quantity)))
	}

	return &ports.CalculateTotalResponse{
		TotalTax:      totalTax,
		TotalDiscount: totalDiscount,
		FinalPrice:    finalPrice,
		Subtotal:      subtotal,
		Policies:      resp.Total.GetPolicies(),
	}, nil
}

// loadTLSCredentials loads TLS credentials from a CA certificate file.
func loadTLSCredentials(certPath string) (credentials.TransportCredentials, error) {
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	if !certPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}
	return credentials.NewTLS(&tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}), nil
}
