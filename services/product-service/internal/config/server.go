package config

import "errors"

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPCPort int
	HTTPPort int
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		GRPCPort: getEnvInt("GRPC_PORT", 50051),
		HTTPPort: getEnvInt("HTTP_PORT", 8080),
	}
}

func (c ServerConfig) Validate() error {
	if c.GRPCPort <= 0 || c.GRPCPort > 65535 {
		return errors.New("invalid gRPC port")
	}
	if c.HTTPPort <= 0 || c.HTTPPort > 65535 {
		return errors.New("invalid HTTP port")
	}
	return nil
}
