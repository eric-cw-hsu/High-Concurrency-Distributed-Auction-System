package config

import "errors"

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPCPort int
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		GRPCPort: getEnvInt("GRPC_PORT", 50051),
	}
}

func (c ServerConfig) Validate() error {
	if c.GRPCPort <= 0 || c.GRPCPort > 65535 {
		return errors.New("invalid gRPC port")
	}
	return nil
}
