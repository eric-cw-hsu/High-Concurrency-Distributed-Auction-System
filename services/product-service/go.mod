module github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/product-service

go 1.25.1

require github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lmittmann/tint v1.1.2 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/samborkent/uuidv7 v0.0.0-20231110121620-f2e19d87e48b // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared => ../shared
