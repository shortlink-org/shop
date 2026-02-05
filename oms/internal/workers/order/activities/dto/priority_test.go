package dto

import (
	"testing"

	"github.com/stretchr/testify/require"

	orderv1 "github.com/shortlink-org/shop/oms/internal/domain/order/v1"
	"github.com/shortlink-org/shop/oms/internal/domain/ports"
)

func Test_domainPriorityToDTO(t *testing.T) {
	t.Run("maps all known domain priorities to DTO", func(t *testing.T) {
		tests := []struct {
			domain   orderv1.DeliveryPriority
			expected ports.DeliveryPriorityDTO
		}{
			{orderv1.DeliveryPriorityUnspecified, ports.DeliveryPriorityUnspecified},
			{orderv1.DeliveryPriorityNormal, ports.DeliveryPriorityNormal},
			{orderv1.DeliveryPriorityUrgent, ports.DeliveryPriorityUrgent},
		}
		for _, tt := range tests {
			got, err := domainPriorityToDTO(tt.domain)
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		}
	})

	t.Run("returns error for unknown priority value", func(t *testing.T) {
		_, err := domainPriorityToDTO(orderv1.DeliveryPriority(99))
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUnsupportedDeliveryPriority, "expected ErrUnsupportedDeliveryPriority")
	})
}
