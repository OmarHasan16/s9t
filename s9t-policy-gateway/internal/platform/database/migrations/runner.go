// internal/platform/database/migrations/runner.go
package migrations

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Runner handles database migrations
type Runner struct {
	dbURL         string
	migrationsDir string
}

// NewRunner creates a new migration runner
func NewRunner(dbURL, migrationsDir string) *Runner {
	return &Runner{
		dbURL:         dbURL,
		migrationsDir: fmt.Sprintf("file://%s", migrationsDir),
	}
}

// Up applies all pending migrations
func (r *Runner) Up() error {
	m, err := migrate.New(r.migrationsDir, r.dbURL)
	if err != nil {
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}
	defer m.Close()

	log.Println("Running database migrations...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database is up to date. No migrations to apply.")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Database migrations applied successfully.")
	return nil
}
