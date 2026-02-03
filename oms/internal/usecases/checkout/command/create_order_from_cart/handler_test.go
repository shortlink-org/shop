package create_order_from_cart

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	cartv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
	itemv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/item/v1"
	itemsv1 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1/items/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/domain/ports/mocks"
)

// mockLogger is a simple mock for the logger interface
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...slog.Attr)                                 {}
func (m *mockLogger) Info(msg string, args ...slog.Attr)                                  {}
func (m *mockLogger) Warn(msg string, args ...slog.Attr)                                  {}
func (m *mockLogger) Error(msg string, args ...slog.Attr)                                 {}
func (m *mockLogger) DebugWithContext(ctx context.Context, msg string, args ...slog.Attr) {}
func (m *mockLogger) InfoWithContext(ctx context.Context, msg string, args ...slog.Attr)  {}
func (m *mockLogger) WarnWithContext(ctx context.Context, msg string, args ...slog.Attr)  {}
func (m *mockLogger) ErrorWithContext(ctx context.Context, msg string, args ...slog.Attr) {}
func (m *mockLogger) Close() error                                                        { return nil }

func TestHandler_Handle_WithPricer(t *testing.T) {
	// Setup
	ctx := context.Background()
	customerID := uuid.New()
	goodID := uuid.New()

	// Create cart with item
	item, err := itemv1.NewItemWithPricing(goodID, 2, decimal.NewFromInt(50), decimal.Zero, decimal.Zero)
	require.NoError(t, err)
	cart := cartv1.Reconstitute(customerID, itemsv1.Items{item}, 1)

	// Create mocks
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockCartRepo := mocks.NewMockCartRepository(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)
	mockPricer := mocks.NewMockPricerClient(t)

	// Setup expectations
	mockUoW.EXPECT().Begin(mock.Anything).Return(ctx, nil)
	mockUoW.EXPECT().Commit(mock.Anything).Return(nil)
	mockUoW.EXPECT().Rollback(mock.Anything).Return(nil)

	mockCartRepo.EXPECT().Load(mock.Anything, customerID).Return(cart, nil)
	mockCartRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockOrderRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

	mockPricer.EXPECT().CalculateTotal(mock.Anything, mock.Anything).Return(
		&ports.CalculateTotalResponse{
			Subtotal:      decimal.NewFromInt(100),
			TotalDiscount: decimal.NewFromInt(10),
			TotalTax:      decimal.NewFromInt(5),
			FinalPrice:    decimal.NewFromInt(95),
			Policies:      []string{"quantity_discount", "vat"},
		},
		nil,
	)

	// Create handler
	handler, err := NewHandler(
		&mockLogger{},
		mockUoW,
		mockCartRepo,
		mockOrderRepo,
		mockPublisher,
		mockPricer,
	)
	require.NoError(t, err)

	// Execute
	cmd := NewCommand(customerID, nil)
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result.Order)
	assert.Equal(t, decimal.NewFromInt(100), result.Subtotal)
	assert.Equal(t, decimal.NewFromInt(10), result.TotalDiscount)
	assert.Equal(t, decimal.NewFromInt(5), result.TotalTax)
	assert.Equal(t, decimal.NewFromInt(95), result.FinalPrice)
}

func TestHandler_Handle_WithoutPricer(t *testing.T) {
	// Test graceful degradation when pricer is nil
	ctx := context.Background()
	customerID := uuid.New()
	goodID := uuid.New()

	// Create cart with item
	item, err := itemv1.NewItemWithPricing(goodID, 2, decimal.NewFromInt(50), decimal.Zero, decimal.Zero)
	require.NoError(t, err)
	cart := cartv1.Reconstitute(customerID, itemsv1.Items{item}, 1)

	// Create mocks
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockCartRepo := mocks.NewMockCartRepository(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	// Setup expectations
	mockUoW.EXPECT().Begin(mock.Anything).Return(ctx, nil)
	mockUoW.EXPECT().Commit(mock.Anything).Return(nil)
	mockUoW.EXPECT().Rollback(mock.Anything).Return(nil)

	mockCartRepo.EXPECT().Load(mock.Anything, customerID).Return(cart, nil)
	mockCartRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockOrderRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

	// Create handler with nil pricer
	handler, err := NewHandler(
		&mockLogger{},
		mockUoW,
		mockCartRepo,
		mockOrderRepo,
		mockPublisher,
		nil, // No pricer client
	)
	require.NoError(t, err)

	// Execute
	cmd := NewCommand(customerID, nil)
	result, err := handler.Handle(ctx, cmd)

	// Assert - should succeed without pricing info
	assert.NoError(t, err)
	assert.NotNil(t, result.Order)
	assert.True(t, result.Subtotal.IsZero())
	assert.True(t, result.TotalDiscount.IsZero())
	assert.True(t, result.TotalTax.IsZero())
	assert.True(t, result.FinalPrice.IsZero())
}

func TestHandler_Handle_PricerError(t *testing.T) {
	// Test handling of pricer errors (graceful degradation)
	ctx := context.Background()
	customerID := uuid.New()
	goodID := uuid.New()

	// Create cart with item
	item, err := itemv1.NewItemWithPricing(goodID, 2, decimal.NewFromInt(50), decimal.Zero, decimal.Zero)
	require.NoError(t, err)
	cart := cartv1.Reconstitute(customerID, itemsv1.Items{item}, 1)

	// Create mocks
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockCartRepo := mocks.NewMockCartRepository(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)
	mockPricer := mocks.NewMockPricerClient(t)

	// Setup expectations
	mockUoW.EXPECT().Begin(mock.Anything).Return(ctx, nil)
	mockUoW.EXPECT().Commit(mock.Anything).Return(nil)
	mockUoW.EXPECT().Rollback(mock.Anything).Return(nil)

	mockCartRepo.EXPECT().Load(mock.Anything, customerID).Return(cart, nil)
	mockCartRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockOrderRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

	// Pricer returns error
	mockPricer.EXPECT().CalculateTotal(mock.Anything, mock.Anything).Return(
		nil,
		errors.New("pricer service unavailable"),
	)

	// Create handler
	handler, err := NewHandler(
		&mockLogger{},
		mockUoW,
		mockCartRepo,
		mockOrderRepo,
		mockPublisher,
		mockPricer,
	)
	require.NoError(t, err)

	// Execute
	cmd := NewCommand(customerID, nil)
	result, err := handler.Handle(ctx, cmd)

	// Assert - should succeed with warning (graceful degradation)
	assert.NoError(t, err)
	assert.NotNil(t, result.Order)
	// Pricing should be zero since pricer failed
	assert.True(t, result.Subtotal.IsZero())
	assert.True(t, result.TotalDiscount.IsZero())
}

func TestHandler_Handle_EmptyCart(t *testing.T) {
	// Test checkout with empty cart
	ctx := context.Background()
	customerID := uuid.New()

	// Create empty cart
	cart := cartv1.New(customerID)

	// Create mocks
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockCartRepo := mocks.NewMockCartRepository(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)

	// Setup expectations
	mockUoW.EXPECT().Begin(mock.Anything).Return(ctx, nil)
	mockUoW.EXPECT().Rollback(mock.Anything).Return(nil)

	mockCartRepo.EXPECT().Load(mock.Anything, customerID).Return(cart, nil)

	// Create handler
	handler, err := NewHandler(
		&mockLogger{},
		mockUoW,
		mockCartRepo,
		mockOrderRepo,
		mockPublisher,
		nil,
	)
	require.NoError(t, err)

	// Execute
	cmd := NewCommand(customerID, nil)
	result, err := handler.Handle(ctx, cmd)

	// Assert - should fail with empty cart error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty cart")
	assert.Nil(t, result.Order)
}

func TestHandler_Handle_MultipleItems(t *testing.T) {
	// Test checkout with multiple items
	ctx := context.Background()
	customerID := uuid.New()
	goodID1 := uuid.New()
	goodID2 := uuid.New()
	goodID3 := uuid.New()

	// Create cart with multiple items
	item1, _ := itemv1.NewItemWithPricing(goodID1, 2, decimal.NewFromInt(25), decimal.Zero, decimal.Zero)
	item2, _ := itemv1.NewItemWithPricing(goodID2, 1, decimal.NewFromInt(50), decimal.Zero, decimal.Zero)
	item3, _ := itemv1.NewItemWithPricing(goodID3, 3, decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
	cart := cartv1.Reconstitute(customerID, itemsv1.Items{item1, item2, item3}, 1)

	// Create mocks
	mockUoW := mocks.NewMockUnitOfWork(t)
	mockCartRepo := mocks.NewMockCartRepository(t)
	mockOrderRepo := mocks.NewMockOrderRepository(t)
	mockPublisher := mocks.NewMockEventPublisher(t)
	mockPricer := mocks.NewMockPricerClient(t)

	// Setup expectations
	mockUoW.EXPECT().Begin(mock.Anything).Return(ctx, nil)
	mockUoW.EXPECT().Commit(mock.Anything).Return(nil)
	mockUoW.EXPECT().Rollback(mock.Anything).Return(nil)

	mockCartRepo.EXPECT().Load(mock.Anything, customerID).Return(cart, nil)
	mockCartRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockOrderRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	mockPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

	// Subtotal: 2*25 + 1*50 + 3*10 = 50 + 50 + 30 = 130
	// Discount: 13 (10%)
	// Tax: 6.5 (5%)
	// Final: 130 - 13 + 6.5 = 123.5
	mockPricer.EXPECT().CalculateTotal(mock.Anything, mock.Anything).Return(
		&ports.CalculateTotalResponse{
			Subtotal:      decimal.NewFromFloat(130),
			TotalDiscount: decimal.NewFromFloat(13),
			TotalTax:      decimal.NewFromFloat(6.5),
			FinalPrice:    decimal.NewFromFloat(123.5),
			Policies:      []string{"combination_discount", "vat"},
		},
		nil,
	)

	// Create handler
	handler, err := NewHandler(
		&mockLogger{},
		mockUoW,
		mockCartRepo,
		mockOrderRepo,
		mockPublisher,
		mockPricer,
	)
	require.NoError(t, err)

	// Execute
	cmd := NewCommand(customerID, nil)
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result.Order)
	assert.Equal(t, decimal.NewFromFloat(130), result.Subtotal)
	assert.Equal(t, decimal.NewFromFloat(13), result.TotalDiscount)
	assert.Equal(t, decimal.NewFromFloat(6.5), result.TotalTax)
	assert.Equal(t, decimal.NewFromFloat(123.5), result.FinalPrice)
}
