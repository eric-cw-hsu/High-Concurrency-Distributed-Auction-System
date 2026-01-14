package config

type Config struct {
	ServiceName string
	Env         string

	GRPC GRPCConfig
	HTTP HTTPConfig
}

type GRPCConfig struct {
	AuthService    GRPCClientConfig
	ProductService GRPCClientConfig
	StockService   GRPCClientConfig
}

type HTTPConfig struct {
	Port string
}

type GRPCClientConfig struct {
	Host string
	Port string
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "api-gateway"),
		Env:         getEnv("ENVIRONMENT", "development"),

		GRPC: GRPCConfig{
			AuthService: GRPCClientConfig{
				Host: getEnv("GRPC_AUTH_SERVICE_HOST", "localhost"),
				Port: getEnv("GRPC_AUTH_SERVICE_PORT", "50051"),
			},
			ProductService: GRPCClientConfig{
				Host: getEnv("GRPC_AUTH_SERVICE_HOST", "localhost"),
				Port: getEnv("GRPC_AUTH_SERVICE_PORT", "50052"),
			},

			StockService: GRPCClientConfig{
				Host: getEnv("STOCK_SERVICE_HOST", "localhost"),
				Port: getEnv("STOCK_SERVICE_PORT", "50053"),
			},
		},

		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
	}
}
