package config

import (
	"fmt"
)

// GRPCConfig aggregates all gRPC related settings for both inbound and outbound traffic.
type GRPCConfig struct {
	// Server settings for the Order Service's own gRPC listener
	Server struct {
		Port int
	}

	// ProductClient settings for connecting to the Product microservice
	ProductClient struct {
		Addr    string
		Timeout int // connection timeout in seconds
	}
}

// loadGRPCConfig initializes gRPC settings from environment variables.
func loadGRPCConfig() GRPCConfig {
	var cfg GRPCConfig

	// Inbound: The port this service will listen on
	cfg.Server.Port = getEnvInt("ORDER_GRPC_PORT", 50051)

	// Outbound: The address and timeout for the external Product Service
	cfg.ProductClient.Addr = getEnv("PRODUCT_SERVICE_ADDR", "localhost:50051")
	cfg.ProductClient.Timeout = getEnvInt("PRODUCT_SERVICE_TIMEOUT", 5)

	return cfg
}

// Validate checks if the gRPC configuration is semantically correct.
func (c *GRPCConfig) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid order_grpc_port: %d", c.Server.Port)
	}

	if c.ProductClient.Addr == "" {
		return fmt.Errorf("product_service_addr must not be empty")
	}

	if c.ProductClient.Timeout <= 0 {
		return fmt.Errorf("product_service_timeout must be a positive integer")
	}

	return nil
}
