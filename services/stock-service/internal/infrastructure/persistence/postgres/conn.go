package postgres

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// MustConnect connects to PostgreSQL or panics
func MustConnect(cfg config.DatabaseConfig) *sqlx.DB {
	zap.L().Info("connecting to postgresql")

	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		zap.L().Fatal("failed to connect to database", zap.Error(err))
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		zap.L().Fatal("failed to ping database", zap.Error(err))
	}

	zap.L().Info("postgresql connected successfully")

	return db
}
