package courier_di

import (
	"context"
	"fmt"

	sdkctx "github.com/shortlink-org/go-sdk/context"
)

func newSDKContext() (context.Context, func(), error) {
	ctx, cancel, err := sdkctx.New()
	if err != nil {
		return nil, nil, fmt.Errorf("sdk context: %w", err)
	}

	return ctx, func() { cancel(nil) }, nil
}
