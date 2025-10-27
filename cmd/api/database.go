package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.maxOpenConns)
	db.SetMaxIdleConns(cfg.DB.maxIdleConns)

	duration, err := time.ParseDuration(cfg.DB.maxIdleTime)
	if err != nil {
		log.Fatal(err)
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}
