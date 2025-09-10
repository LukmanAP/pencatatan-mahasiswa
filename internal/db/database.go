package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool adalah alias untuk pgxpool.Pool agar tipe dependensi seragam di seluruh proyek
// Dengan ini, *db.Pool identik dengan *pgxpool.Pool
// sehingga fungsi yang menerima *db.Pool bisa menerima hasil dari Connect()
type Pool = pgxpool.Pool

func Connect(url string) (*pgxpool.Pool, error) {
	// Hindari mencetak URL database secara penuh karena berpotensi mengandung kredensial

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Printf("Error parsing database config: %v", err)
		return nil, err
	}

	cfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Printf("Error creating connection pool: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("Error pinging database: %v", err)
		pool.Close()
		return nil, err
	}

	log.Printf("Successfully connected to database")
	return pool, nil
}
