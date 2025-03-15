# PROTO TASKS ==========================================================================================================

proto: proto-lock proto-lint proto-generate ## Check lint, lock proto dependencies and generate proto-files

proto-lint: ## Check lint
	@buf ls-files
	@buf lint
	@buf breaking --against ops/proto/proto-lock.json

proto-lock: ## Lock proto dependencies
	@buf build -o ops/proto/proto-lock.json

proto-generate: ## Generate proto-files
	# domain --------------------------------------------------------------------------------------
	@buf generate \
		--path=internal/domain \
		--template=ops/proto/domain.buf.gen.yaml

    # rpc -----------------------------------------------------------------------------------------
	@buf generate \
		--path=internal/infrastructure \
		--template=ops/proto/rpc.buf.gen.yaml

	# workers -------------------------------------------------------------------------------------
	@buf generate \
		--path=internal/workers \
		--template=ops/proto/domain.buf.gen.yaml
