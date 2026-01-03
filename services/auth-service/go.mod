module github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service

go 1.25.1

require github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/redis/go-redis/v9 v9.17.2 // indirect
	github.com/samborkent/uuidv7 v0.0.0-20231110121620-f2e19d87e48b // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/shared => ../shared
