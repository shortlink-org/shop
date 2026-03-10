package dto

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"

	domain "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/rpcmeta"
	model "github.com/shortlink-org/shop/oms/internal/infrastructure/rpc/cart/v1/model/v1"
)

func TestAddRequestToDomain(t *testing.T) {
	validCustomerID := "e2c8ba97-1a6b-4c5c-9a2a-3f4c9b9d65a1"
	ctxValid := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", validCustomerID))
	ctxInvalidCustomer := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", "invalid-uuid"))
	ctxMissing := context.Background()

	tests := []struct {
		name          string
		ctx           context.Context
		request       *model.AddRequest
		expectedError error
		expectedState *domain.State
	}{
		{
			name: "Valid AddRequest",
			ctx:  ctxValid,
			request: &model.AddRequest{
				Items: []*model.CartItem{
					{GoodId: "c5f5d6d6-98e6-4f57-b34a-48a3997f28d4", Quantity: 1},
					{GoodId: "da3f3a3e-784d-4a9a-8cfa-6321d555d6a3", Quantity: 2},
				},
			},
			expectedError: nil,
			expectedState: func() *domain.State {
				customerId, err := uuid.Parse(validCustomerID)
				if err != nil {
					panic(err)
				}
				goodId1, err := uuid.Parse("c5f5d6d6-98e6-4f57-b34a-48a3997f28d4")
				if err != nil {
					panic(err)
				}
				goodId2, err := uuid.Parse("da3f3a3e-784d-4a9a-8cfa-6321d555d6a3")
				if err != nil {
					panic(err)
				}
				state := domain.New(customerId)
				item1, err := itemv1.NewItem(goodId1, 1)
				if err != nil {
					panic(err)
				}
				if addErr := state.AddItem(item1); addErr != nil {
					panic(addErr)
				}
				item2, err := itemv1.NewItem(goodId2, 2)
				if err != nil {
					panic(err)
				}
				if addErr := state.AddItem(item2); addErr != nil {
					panic(addErr)
				}

				return state
			}(),
		},
		{
			name:          "Invalid Customer ID (x-user-id)",
			ctx:           ctxInvalidCustomer,
			request:       &model.AddRequest{Items: []*model.CartItem{{GoodId: uuid.New().String(), Quantity: 1}}},
			expectedError: rpcmeta.ErrMissingCustomerID,
			expectedState: nil,
		},
		{
			name:          "Missing x-user-id",
			ctx:           ctxMissing,
			request:       &model.AddRequest{Items: []*model.CartItem{{GoodId: uuid.New().String(), Quantity: 1}}},
			expectedError: rpcmeta.ErrMissingCustomerID,
			expectedState: nil,
		},
		{
			name: "Invalid Good ID",
			ctx:  ctxValid,
			request: &model.AddRequest{
				Items: []*model.CartItem{
					{GoodId: "invalid-uuid", Quantity: 1},
				},
			},
			expectedError: ParseItemError{
				Err:  errors.New("invalid UUID length: 12"),
				item: "invalid-uuid",
			},
			expectedState: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualParams, err := AddRequestToDomain(tt.ctx, tt.request)
			if tt.expectedError != nil {
				assert.Error(t, err)
				if errors.Is(err, rpcmeta.ErrMissingCustomerID) {
					assert.True(t, errors.Is(err, tt.expectedError), "got %v", err)
				} else {
					var parseErr ParseItemError
					assert.True(t, errors.As(err, &parseErr), "expected ParseItemError, got %v", err)
					assert.Equal(t, tt.expectedError.(ParseItemError).item, parseErr.item)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actualParams)
				assert.Equal(t, tt.expectedState.GetCustomerId(), actualParams.CustomerID)
				assert.Len(t, actualParams.Items, len(tt.expectedState.GetItems()))
			}
		})
	}
}
