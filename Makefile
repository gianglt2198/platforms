

folder: ## create folder structure
	@mkdir -p cmd
	@mkdir -p cmd/server
	@mkdir -p cmd/client
	@mkdir -p config
	@mkdir -p internal
	@mkdir -p internal/app
	@mkdir -p internal/config
	@mkdir -p internal/core
	@mkdir -p internal/services/rest
	@mkdir -p internal/services/grpc
	@mkdir -p internal/services/graphql
	@mkdir -p internal/services/ed
	@mkdir -p internal/tools/nats
	@mkdir -p internal/tools/database
	@mkdir -p internal/tools/redis
	@mkdir -p pkg/logger
	@mkdir -p pkg/utils
