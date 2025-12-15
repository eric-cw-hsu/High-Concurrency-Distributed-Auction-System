package envloader

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// init loads .env files in a prioritized order so packages that rely on
// environment variables during their init() will see them. Preferred simple
// pattern for easy mapping to K8s ConfigMaps:
//  1. .env (repo root)
//  2. .env.<SERVICE_NAME> (repo root) where SERVICE_NAME is taken from
//     the SERVICE_NAME or LOGGER_SERVICE_NAME env var (or inferred)
func init() {
	// always attempt to load root .env first
	if err := godotenv.Load(); err == nil {
		log.Printf("envloader: loaded .env from repo root")
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName != "" {
		if err := godotenv.Overload(".env." + serviceName); err == nil {
			log.Printf("envloader: loaded .env.%s from repo root", serviceName)
		}
	}
}
