package postgres

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/order-service/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func MustConnect(cfg config.DatabaseConfig) *sqlx.DB {
	db, err := sqlx.Connect(cfg.Driver, cfg.DSN)
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}

	// Connection pool tuning
	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.Lifetime)

	if err := db.Ping(); err != nil {
		zap.L().Fatal("database ping failed", zap.Error(err))
	}

	return db
}
