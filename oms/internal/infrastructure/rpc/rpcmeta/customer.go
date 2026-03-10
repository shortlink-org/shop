package rpcmeta

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// XUserIDKey is the gRPC metadata key for the authenticated user (set by Istio from JWT).
const XUserIDKey = "x-user-id"

// ErrMissingCustomerID is returned when x-user-id is missing or invalid (caller should return gRPC Unauthenticated).
var ErrMissingCustomerID = errors.New("missing or invalid customer identity (x-user-id)")

// CustomerIDFromContext returns the customer UUID from incoming gRPC metadata (x-user-id).
// Istio RequestAuthentication with outputClaimToHeaders injects it from the JWT sub claim.
func CustomerIDFromContext(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, fmt.Errorf("%w: missing metadata", ErrMissingCustomerID)
	}

	vals := md.Get(XUserIDKey)
	if len(vals) == 0 || vals[0] == "" {
		return uuid.Nil, fmt.Errorf("%w: missing %s", ErrMissingCustomerID, XUserIDKey)
	}

	id, err := uuid.Parse(vals[0])
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid %s: %w", ErrMissingCustomerID, XUserIDKey, err)
	}

	return id, nil
}
