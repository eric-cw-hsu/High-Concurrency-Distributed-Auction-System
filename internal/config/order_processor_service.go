package config

type OrderProcessorServiceConfig struct {
	Name        string
	MetricsPort int
}

func LoadOrderProcessorServiceConfig() *OrderProcessorServiceConfig {
	return &OrderProcessorServiceConfig{
		Name:        getEnv("ORDER_PROCESSOR_SERVICE_NAME", "order-service"),
		MetricsPort: getEnvAsInt("ORDER_PROCESSOR_SERVICE_METRICS_PORT", 9100),
	}
}
