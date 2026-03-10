package dto

import (
	commonpb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/oms/domain/order/v1/common"
	servicepb "github.com/shortlink-org/shop/oms-graphql/pkg/generated/service/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// RecipientContactsToService maps OMS recipient contacts to Connect response.
func RecipientContactsToService(contacts *commonpb.RecipientContacts) *servicepb.RecipientContacts {
	if contacts == nil {
		return nil
	}
	return &servicepb.RecipientContacts{
		RecipientName:  wrapperspb.String(contacts.GetRecipientName()),
		RecipientPhone: wrapperspb.String(contacts.GetRecipientPhone()),
		RecipientEmail: wrapperspb.String(contacts.GetRecipientEmail()),
	}
}

// RecipientContactsFromInput maps Connect recipient contacts input to OMS proto.
func RecipientContactsFromInput(input *servicepb.RecipientContactsInput) *commonpb.RecipientContacts {
	if input == nil {
		return nil
	}
	contacts := &commonpb.RecipientContacts{}
	if input.GetRecipientName() != nil {
		contacts.RecipientName = input.GetRecipientName().GetValue()
	}
	if input.GetRecipientPhone() != nil {
		contacts.RecipientPhone = input.GetRecipientPhone().GetValue()
	}
	if input.GetRecipientEmail() != nil {
		contacts.RecipientEmail = input.GetRecipientEmail().GetValue()
	}
	if contacts.RecipientName == "" && contacts.RecipientPhone == "" && contacts.RecipientEmail == "" {
		return nil
	}
	return contacts
}
