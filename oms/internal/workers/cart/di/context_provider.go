package oms_cart_worker_di

import (
	"context"

	sdkctx "github.com/shortlink-org/go-sdk/context"
)

func newSDKContext() (context.Context, func(), error) {
	ctx, cancel, err := sdkctx.New()
	if err != nil {
		return nil, nil, err
	}

	return ctx, func() { cancel(nil) }, nil
}
