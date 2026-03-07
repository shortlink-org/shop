package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestQueryGetCartRequiresUserID(t *testing.T) {
	t.Parallel()

	svc := New(nil, nil, nil)
	req := connect.NewRequest(&servicepb.QueryGetCartRequest{})

	_, err := svc.QueryGetCart(context.Background(), req)
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}

	if code := connect.CodeOf(err); code != connect.CodeUnauthenticated {
		t.Fatalf("expected code %v, got %v", connect.CodeUnauthenticated, code)
	}
}

func TestMapCreateOrderRequestUsesUserIDAndGeneratesOrderID(t *testing.T) {
	t.Parallel()

	req, err := mapCreateOrderRequest("user-123", &servicepb.CreateOrderInput{
		Items: []*servicepb.OrderItemInput{
			{
				Id:       "good-1",
				Quantity: 2,
				Price:    99.5,
			},
		},
	})
	if err != nil {
		t.Fatalf("mapCreateOrderRequest returned error: %v", err)
	}

	if req.GetOrder().GetCustomerId() != "user-123" {
		t.Fatalf("expected customer id user-123, got %q", req.GetOrder().GetCustomerId())
	}

	if req.GetOrder().GetId() == "" {
		t.Fatal("expected generated order id")
	}

	if len(req.GetOrder().GetItems()) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(req.GetOrder().GetItems()))
	}
}
