package v1

import (
	v2 "github.com/shortlink-org/shop/oms/internal/domain/cart/v1"
)

type CartEvent struct {
	Event v2.Event
	Items []*WorkerCartItem
}
