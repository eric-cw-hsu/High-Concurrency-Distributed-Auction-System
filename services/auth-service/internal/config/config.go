package config

import "time"

type Config struct {
	ServiceName string
	Env         string

	GRPC     GRPCConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Bcrypt   BcryptConfig
	Logger   LoggerConfig
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type GRPCConfig struct {
	Port string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecretKey  string
	RefreshSecretKey string
	AccessTTL        int // in seconds
	RefreshTTL       int // in seconds
	Issuer           string
	Audience         string
}

type BcryptConfig struct {
	Cost int
}

func Load() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "auth-service"),
		Env:         getEnv("ENV", "local"),

		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "50051"),
		},

		Database: DatabaseConfig{
			DSN:             getEnv("DB_DSN", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN", 50),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE", 10),
			ConnMaxLifetime: time.Minute * 5,
		},

		JWT: JWTConfig{
			AccessSecretKey:  mustEnv("JWT_ACCESS_SECRET_KEY"),
			RefreshSecretKey: mustEnv("JWT_REFRESH_SECRET_KEY"),
			AccessTTL:        getEnvAsInt("JWT_ACCESS_TTL", 15*60),    // default 15 minutes
			RefreshTTL:       getEnvAsInt("JWT_REFRESH_TTL", 1440*60), // default 1440 minutes (1 day)
			Issuer:           getEnv("JWT_ISSUER", "auth-service"),
			Audience:         getEnv("JWT_AUDIENCE", "auth-clients"),
		},

		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},

		Bcrypt: BcryptConfig{
			Cost: getEnvAsInt("BCRYPT_COST", 10),
		},

		Logger: loadLoggerConfig(),
	}
}
