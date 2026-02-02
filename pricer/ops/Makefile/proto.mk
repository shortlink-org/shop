# PROTO TASKS ==========================================================================================================

generate: proto-generate wire-generate ### Generate proto and wire code

proto-generate: ### Generate Go code from proto files
	@buf generate --path internal/infrastructure/rpc --template ops/proto/buf.gen.yaml

wire-generate: ### Generate wire DI
	@go generate -tags=wireinject ./...
