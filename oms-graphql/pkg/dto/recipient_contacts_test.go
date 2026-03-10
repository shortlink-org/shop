package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"

	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
)

func TestRecipientContactsToService(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, RecipientContactsToService(nil))
	})
	t.Run("maps contacts", func(t *testing.T) {
		in := &commonpb.RecipientContacts{
			RecipientName: "n", RecipientPhone: "p", RecipientEmail: "e",
		}
		out := RecipientContactsToService(in)
		assert.Equal(t, "n", out.RecipientName.GetValue())
		assert.Equal(t, "p", out.RecipientPhone.GetValue())
	})
}

func TestRecipientContactsFromInput(t *testing.T) {
	t.Parallel()
	t.Run("nil returns nil", func(t *testing.T) {
		assert.Nil(t, RecipientContactsFromInput(nil))
	})
	t.Run("empty contacts returns nil", func(t *testing.T) {
		out := RecipientContactsFromInput(&servicepb.RecipientContactsInput{})
		assert.Nil(t, out)
	})
	t.Run("maps input", func(t *testing.T) {
		in := &servicepb.RecipientContactsInput{
			RecipientName:  wrapperspb.String("name"),
			RecipientPhone: wrapperspb.String("+1"),
		}
		out := RecipientContactsFromInput(in)
		assert.Equal(t, "name", out.RecipientName)
		assert.Equal(t, "+1", out.RecipientPhone)
	})
}
