package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
)

const (
	authHeader        = "Authorization"
	xUserIDHeader     = "X-User-ID"
	traceparentHeader = "traceparent"
	traceIDHeader     = "trace-id"
)

var errMissingTrackingIdentity = errors.New("missing required Authorization or X-User-ID header")

func trackingContextFromHeaders(ctx context.Context, headers http.Header) (context.Context, error) {
	if strings.TrimSpace(headers.Get(authHeader)) == "" && strings.TrimSpace(headers.Get(xUserIDHeader)) == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errMissingTrackingIdentity)
	}

	return forwardContextFromHeaders(ctx, headers), nil
}

func forwardContextFromHeaders(ctx context.Context, headers http.Header) context.Context {
	mdPairs := make([]string, 0, 8)

	for _, key := range []string{authHeader, xUserIDHeader, traceparentHeader, traceIDHeader} {
		if value := strings.TrimSpace(headers.Get(key)); value != "" {
			mdPairs = append(mdPairs, strings.ToLower(key), value)
		}
	}

	if len(mdPairs) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, metadata.Pairs(mdPairs...))
}
