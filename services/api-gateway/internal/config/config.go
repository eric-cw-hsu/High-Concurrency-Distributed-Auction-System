package config

type Config struct {
	ServiceName string
	Env         string

	GRPC GRPCConfig
	HTTP HTTPConfig
}

type GRPCConfig struct {
	AuthService GRPCAuthConfig
}

type HTTPConfig struct {
	Port string
}

type GRPCAuthConfig struct {
	Host string
	Port string
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "api-gateway"),
		Env:         getEnv("ENVIRONMENT", "development"),

		GRPC: GRPCConfig{
			AuthService: GRPCAuthConfig{
				Host: getEnv("GRPC_AUTH_SERVICE_HOST", "localhost"),
				Port: getEnv("GRPC_AUTH_SERVICE_PORT", "50051"),
			},
		},

		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
	}
}
