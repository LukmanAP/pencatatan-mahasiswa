package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// RunMigrations membaca file migrasi dari folder "migrations" di working directory
// dan menjalankannya terhadap database sesuai databaseURL.
// Catatan: fungsi ini akan memanggil log.Fatalf pada error fatal.
func RunMigrations(databaseURL string) {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working dir: %v", err)
	}
	migDir := filepath.Join(wd, "migrations")

	// Gunakan source iofs agar kompatibel lintas OS (khususnya Windows)
	d, err := iofs.New(os.DirFS(wd), "migrations")
	if err != nil {
		log.Fatalf("failed to init iofs source for migrations at %s: %v", migDir, err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("migration: no changes")
			return
		}
		log.Fatalf("migration failed: %v", err)
		return
	}
	log.Println("migration: applied successfully")
}
