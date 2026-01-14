.PHONY: proto clean

proto:
	@echo "Generating proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/product/v1/*.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/stock/v1/*.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auth/v1/*.proto
	@echo "Proto generation complete"

clean:
	find proto -name "*.pb.go" -delete
	@echo "Cleaned generated proto files"