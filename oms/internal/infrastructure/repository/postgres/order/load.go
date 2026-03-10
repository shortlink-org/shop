package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/shortlink-org/shop/oms/internal/domain"
	order "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/dto"
	"github.com/shortlink-org/shop/oms/internal/infrastructure/repository/postgres/order/schema/queries"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

func cloneOrderDeliveryInfo(info *order.DeliveryInfo) *order.DeliveryInfo {
	if info == nil {
		return nil
	}

	var recipientContacts *order.RecipientContacts

	if rc := info.GetRecipientContacts(); rc != nil {
		clone := order.NewRecipientContacts(rc.GetName(), rc.GetPhone(), rc.GetEmail())
		recipientContacts = &clone
	}

	cloned := order.NewDeliveryInfo(
		info.GetPickupAddress(),
		info.GetDeliveryAddress(),
		info.GetDeliveryPeriod(),
		info.GetPackageInfo(),
		info.GetPriority(),
		recipientContacts,
	)

	if packageID := info.GetPackageId(); packageID != nil {
		cloned.SetPackageId(*packageID)
	}

	return &cloned
}

func cloneTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}

func cloneOrderState(state *order.OrderState) *order.OrderState {
	if state == nil {
		return nil
	}

	return order.NewOrderStateFromPersisted(
		state.GetOrderID(),
		state.GetCustomerId(),
		state.GetItems(),
		state.GetStatus(),
		state.GetVersion(),
		cloneOrderDeliveryInfo(state.GetDeliveryInfo()),
		state.GetDeliveryStatus(),
		cloneTimePointer(state.GetDeliveryRequestedAt()),
	)
}

func (s *Store) loadOrderAggregate(ctx context.Context, qtx *queries.Queries, row queries.OmsOrder) (*order.OrderState, error) {
	items, err := qtx.GetOrderItems(ctx, row.ID)
	if err != nil {
		return nil, domain.WrapUnavailable("GetOrderItems", err)
	}

	var deliveryInfoRow *queries.GetOrderDeliveryInfoRow

	deliveryRow, err := qtx.GetOrderDeliveryInfo(ctx, row.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.WrapUnavailable("GetOrderDeliveryInfo", err)
		}
	} else {
		deliveryInfoRow = &deliveryRow
	}

	result := (&dto.OrderRow{Order: row, Items: items, Delivery: deliveryInfoRow}).ToDomain()

	cost := int64(200 + len(items)*50) //nolint:mnd // ristretto cost formula
	s.cache.SetWithTTL(row.ID.String(), cloneOrderState(result), cost, cacheTTL)

	return result, nil
}

// Load retrieves an order by ID.
// Uses L1 cache for frequently accessed orders.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) Load(ctx context.Context, orderID uuid.UUID) (*order.OrderState, error) {
	// Check L1 cache first
	cacheKey := orderID.String()
	if cachedOrder, found := s.cache.Get(cacheKey); found {
		return cloneOrderState(cachedOrder), nil
	}

	// Cache miss - fetch from database
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	// Get order header
	row, err := qtx.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}

		return nil, domain.WrapUnavailable("GetOrder", err)
	}

	return s.loadOrderAggregate(ctx, qtx, row)
}

// LoadByPackageID retrieves an order by delivery package ID.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) LoadByPackageID(ctx context.Context, packageID uuid.UUID) (*order.OrderState, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)
	row, err := qtx.GetOrderByPackageID(ctx, pgtype.UUID{Bytes: packageID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrNotFound
		}

		return nil, domain.WrapUnavailable("GetOrderByPackageID", err)
	}

	return s.loadOrderAggregate(ctx, qtx, row)
}

// ListByCustomer retrieves all orders for a customer.
// Requires transaction in context (use UnitOfWork.Begin()).
func (s *Store) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*order.OrderState, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return nil, ErrTransactionRequired
	}

	qtx := s.query.WithTx(pgxTx)

	rows, err := qtx.ListOrdersByCustomer(ctx, customerID)
	if err != nil {
		return nil, domain.WrapUnavailable("ListOrdersByCustomer", err)
	}

	orders := make([]*order.OrderState, 0, len(rows))
	for _, row := range rows {
		// Get items for each order
		items, err := qtx.GetOrderItems(ctx, row.ID)
		if err != nil {
			return nil, domain.WrapUnavailable("GetOrderItems", err)
		}

		// Get delivery info (optional)
		var deliveryInfoRow *queries.GetOrderDeliveryInfoRow

		deliveryRow, err := qtx.GetOrderDeliveryInfo(ctx, row.ID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.WrapUnavailable("GetOrderDeliveryInfo", err)
			}
		} else {
			deliveryInfoRow = &deliveryRow
		}

		orders = append(orders, (&dto.OrderRow{Order: row, Items: items, Delivery: deliveryInfoRow}).ToDomain())
	}

	return orders, nil
}
