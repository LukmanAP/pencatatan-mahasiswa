package main

import (
	"log"
	"pencatatan-data-mahasiswa/api/http"
	"pencatatan-data-mahasiswa/internal/config"
	"pencatatan-data-mahasiswa/internal/db"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()

	r := http.NewRouter()

	r.Run(":" + cfg.AppPort)
}
