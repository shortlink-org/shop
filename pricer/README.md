# Pricer service

> Service for calculating tax and discount for cart/order. Implemented in **Go** with OPA, DDD, and gRPC.

## Status

**Active.** Go implementation with OPA (Rego) policy engine.

## Discount rules (simplified)

Only quantity-based and combination-based discounts:

- **Quantity discount** — e.g. "3 for 2": buy 3 get 1 free (`min_quantity_for_discount`)
- **Combination discount** — 5% off when cart has 2+ different products (`combination_discount_percent`)

No brand-based or time-based rules — input needs only `productId`, `quantity`, `price` per item.

## Stack

- **Go** — implementation language
- **OPA (Rego)** — policy engine for discounts and taxes
- **DDD** — domain/usecases/infrastructure layers
- **Command pattern** — usecases/cart/command/calculate_total
- **gRPC + Buf** — API (CartService.CalculateTotal)
- **go-sdk** — config, logger, observability, gRPC server
- **Wire** — dependency injection

## Modes

- **gRPC mode** (default): Run gRPC server. Set `GRPC_SERVER_ENABLED=false` to disable.
- **CLI mode**: When gRPC is disabled, processes cart files from `cart_files` config.

## Configuration

See `config.yaml` for policy paths, queries, cart files, and output directory.

## Development

```bash
# Generate proto and wire
make generate

# Build
go build ./...
```
