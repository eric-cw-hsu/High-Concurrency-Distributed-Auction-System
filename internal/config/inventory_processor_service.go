package config

type InventoryProcessorServiceConfig struct {
	Name        string
	MetricsPort int
}

func LoadInventoryProcessorServiceConfig() *InventoryProcessorServiceConfig {
	return &InventoryProcessorServiceConfig{
		Name:        getEnv("INVENTORY_PROCESSOR_SERVICE_NAME", "inventory-processor-service"),
		MetricsPort: getEnvAsInt("INVENTORY_PROCESSOR_SERVICE_METRICS_PORT", 9200),
	}
}
