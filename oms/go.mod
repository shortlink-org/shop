module github.com/shortlink-org/shop/oms

go 1.25.6

require (
	github.com/authzed/authzed-go v1.7.0
	github.com/google/uuid v1.6.0
	github.com/google/wire v0.7.0
	github.com/gorilla/websocket v1.5.3
	github.com/shopspring/decimal v1.4.0
	github.com/shortlink-org/go-sdk/auth v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/config v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/context v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/flags v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/fsm v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/graceful_shutdown v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/grpc v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/logger v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/observability v0.0.0-20260130191740-6ca93b747e53
	github.com/shortlink-org/go-sdk/temporal v0.0.0-20260130191740-6ca93b747e53
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel/trace v1.39.0
	go.temporal.io/sdk v1.39.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Unleash/unleash-go-sdk/v5 v5.0.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/fgprof v0.9.5 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/google/pprof v0.0.0-20251213031049-b05bdaca462f // indirect
	github.com/grafana/otel-profiling-go v0.5.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jzelinskie/stringz v0.0.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nexus-rpc/sdk-go v0.5.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20241121165744-79df5c4772f2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.4 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.19.2 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/samber/lo v1.52.0 // indirect
	github.com/shortlink-org/go-sdk/flight_trace v0.0.0-20260130191740-6ca93b747e53 // indirect
	github.com/shortlink-org/go-sdk/http v0.0.0-20260130191740-6ca93b747e53 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.64.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.39.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.39.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.61.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.temporal.io/api v1.59.0 // indirect
	go.temporal.io/sdk/contrib/opentelemetry v0.6.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Local development - replace with local paths when developing
// Comment out for CI/CD or production builds
replace (
	github.com/shortlink-org/go-sdk/auth => ../../go-sdk/auth
	github.com/shortlink-org/go-sdk/config => ../../go-sdk/config
	github.com/shortlink-org/go-sdk/context => ../../go-sdk/context
	github.com/shortlink-org/go-sdk/flags => ../../go-sdk/flags
	github.com/shortlink-org/go-sdk/flight_trace => ../../go-sdk/flight_trace
	github.com/shortlink-org/go-sdk/fsm => ../../go-sdk/fsm
	github.com/shortlink-org/go-sdk/graceful_shutdown => ../../go-sdk/graceful_shutdown
	github.com/shortlink-org/go-sdk/grpc => ../../go-sdk/grpc
	github.com/shortlink-org/go-sdk/http => ../../go-sdk/http
	github.com/shortlink-org/go-sdk/logger => ../../go-sdk/logger
	github.com/shortlink-org/go-sdk/observability => ../../go-sdk/observability
	github.com/shortlink-org/go-sdk/temporal => ../../go-sdk/temporal
)
