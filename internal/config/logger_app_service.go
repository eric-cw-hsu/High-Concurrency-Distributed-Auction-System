package config

type LoggerAppServiceConfig struct {
	Name        string
	MetricsPort int

	LogStorageType string
	LogDirPath     string
}

func LoadLoggerAppServiceConfig() LoggerAppServiceConfig {
	return LoggerAppServiceConfig{
		Name:        getEnv("LOGGER_SERVICE_NAME", "logger-service"),
		MetricsPort: getEnvAsInt("LOGGER_SERVICE_METRICS_PORT", 8082),

		LogStorageType: getEnv("LOG_STORAGE_TYPE", "file"),
		LogDirPath:     getEnv("LOG_DIR_PATH", "logs/"),
	}
}
