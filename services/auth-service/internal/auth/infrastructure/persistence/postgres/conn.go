package postgres

import (
	"database/sql"

	"go.uber.org/zap"
)

func MustConnect(dsn string) *sql.DB {
	zap.L().Info("connecting to postgresql")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		zap.L().Fatal("failed to open database connection",
			zap.Error(err),
		)
	}

	if err := db.Ping(); err != nil {
		zap.L().Fatal("failed to ping database",
			zap.Error(err),
		)
	}

	zap.L().Info("postgresql connected successfully")

	return db
}
