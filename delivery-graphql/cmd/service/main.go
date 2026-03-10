package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/shortlink-org/shop/delivery-graphql/pkg/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	err := service.Start(ctx)
	if err != nil {
		stop()
		log.Fatal(err)
	}

	stop()
}
