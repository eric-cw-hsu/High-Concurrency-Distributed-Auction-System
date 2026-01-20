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
	OrderService   GRPCClientConfig
}

type HTTPConfig struct {
	Port string
}

type GRPCClientConfig struct {
	Host                string
	Port                string
	KeepaliveTime       int // seconds
	KeepaliveTimeout    int // seconds
	PermitWithoutStream bool
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "api-gateway"),
		Env:         getEnv("ENVIRONMENT", "development"),

		GRPC: GRPCConfig{
			AuthService: GRPCClientConfig{
				Host:                getEnv("GRPC_AUTH_SERVICE_HOST", "localhost"),
				Port:                getEnv("GRPC_AUTH_SERVICE_PORT", "50051"),
				KeepaliveTime:       getEnvInt("GRPC_AUTH_SERVICE_KEEPALIVE_TIME", 30),
				KeepaliveTimeout:    getEnvInt("GRPC_AUTH_SERVICE_KEEPALIVE_TIMEOUT", 5),
				PermitWithoutStream: getEnvBool("GRPC_AUTH_SERVICE_PERMIT_WITHOUT_STREAM", false),
			},
			ProductService: GRPCClientConfig{
				Host:                getEnv("GRPC_PRODUCT_SERVICE_HOST", "localhost"),
				Port:                getEnv("GRPC_PRODUCT_SERVICE_PORT", "50052"),
				KeepaliveTime:       getEnvInt("GRPC_PRODUCT_SERVICE_KEEPALIVE_TIME", 30),
				KeepaliveTimeout:    getEnvInt("GRPC_PRODUCT_SERVICE_KEEPALIVE_TIMEOUT", 5),
				PermitWithoutStream: getEnvBool("GRPC_PRODUCT_SERVICE_PERMIT_WITHOUT_STREAM", false),
			},
			StockService: GRPCClientConfig{
				Host:                getEnv("GRPC_STOCK_SERVICE_HOST", "localhost"),
				Port:                getEnv("GRPC_STOCK_SERVICE_PORT", "50053"),
				KeepaliveTime:       getEnvInt("GRPC_STOCK_SERVICE_KEEPALIVE_TIME", 30),
				KeepaliveTimeout:    getEnvInt("GRPC_STOCK_SERVICE_KEEPALIVE_TIMEOUT", 5),
				PermitWithoutStream: getEnvBool("GRPC_STOCK_SERVICE_PERMIT_WITHOUT_STREAM", false),
			},
			OrderService: GRPCClientConfig{
				Host:                getEnv("GRPC_ORDER_SERVICE_HOST", "localhost"),
				Port:                getEnv("GRPC_ORDER_SERVICE_PORT", "50054"),
				KeepaliveTime:       getEnvInt("GRPC_ORDER_SERVICE_KEEPALIVE_TIME", 30),
				KeepaliveTimeout:    getEnvInt("GRPC_ORDER_SERVICE_KEEPALIVE_TIMEOUT", 5),
				PermitWithoutStream: getEnvBool("GRPC_ORDER_SERVICE_PERMIT_WITHOUT_STREAM", false),
			},
		},

		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
	}
}
