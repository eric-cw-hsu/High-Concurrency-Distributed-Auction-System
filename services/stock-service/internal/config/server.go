package config

import (
	"fmt"
	"time"
)

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	GRPCPort          int
	MaxConnectionIdle time.Duration
	MaxConnectionAge  time.Duration
	Timeout           time.Duration
}

// loadServerConfig loads server configuration
func loadServerConfig() ServerConfig {
	return ServerConfig{
		GRPCPort:          getEnvInt("SERVER_GRPC_PORT", 50051),
		MaxConnectionIdle: getEnvDuration("SERVER_MAX_CONNECTION_IDLE", 5*time.Minute),
		MaxConnectionAge:  getEnvDuration("SERVER_MAX_CONNECTION_AGE", 10*time.Minute),
		Timeout:           getEnvDuration("SERVER_TIMEOUT", 30*time.Second),
	}
}

// Validate validates server configuration
func (c *ServerConfig) Validate() error {
	if c.GRPCPort <= 0 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid grpc port: %d", c.GRPCPort)
	}
	return nil
}
