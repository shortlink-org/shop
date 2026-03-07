package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/shortlink-org/shop/admin-graphql/pkg/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := service.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
