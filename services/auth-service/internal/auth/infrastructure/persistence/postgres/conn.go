package postgres

import "database/sql"

func MustConnect(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return db
}
