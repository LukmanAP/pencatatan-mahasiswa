package config

import (
	"log"
	"net"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort     string
	DatabaseURL string
}

// buildDatabaseURLFromEnv merakit connection string Postgres dari variabel env terpisah
func buildDatabaseURLFromEnv() string {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_NAME")

	sslmode := os.Getenv("DATABASE_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Jika variabel utama ada yang kosong, kembalikan string kosong untuk memicu error handling di Load()
	if host == "" || port == "" || user == "" || name == "" {
		return ""
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + name,
	}
	q := url.Values{}
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String()
}

func Load() *Config {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	urlStr := os.Getenv("DATABASE_URL")
	if urlStr == "" {
		urlStr = buildDatabaseURLFromEnv()
	}
	if urlStr == "" {
		log.Fatal("DATABASE_URL or DATABASE_* environment variables are required")
	}
	return &Config{AppPort: port, DatabaseURL: urlStr}
}
