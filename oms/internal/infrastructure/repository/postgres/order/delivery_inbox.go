package postgres

import (
	"context"
	"fmt"

	"github.com/shortlink-org/shop/oms/internal/domain"
	"github.com/shortlink-org/shop/oms/pkg/uow"
)

// TryRecord stores an inbound delivery message once per consumer.
// Returns false when the same message was already committed before.
func (s *Store) TryRecord(ctx context.Context, consumerName, messageID, topic string) (bool, error) {
	pgxTx := uow.FromContext(ctx)
	if pgxTx == nil {
		return false, ErrTransactionRequired
	}

	tag, err := pgxTx.Exec(
		ctx,
		`INSERT INTO oms.delivery_inbox_messages (consumer_name, message_id, topic)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (consumer_name, message_id) DO NOTHING`,
		consumerName,
		messageID,
		topic,
	)
	if err != nil {
		return false, domain.WrapUnavailable("InsertDeliveryInboxMessage", fmt.Errorf("exec insert: %w", err))
	}

	return tag.RowsAffected() == 1, nil
}
