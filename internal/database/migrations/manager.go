package migrations

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func NewMigrateInstance(db *sql.DB) (*migrate.Migrate, error) {
	config, err := utils.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return nil, fmt.Errorf("error creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+config.SchemaPath,
		"pgx5", driver)
	if err != nil {
		return nil, fmt.Errorf("error creating migration instance: %w", err)
	}

	return m, nil
}

func ShouldRunMigrations() bool {
	// Check CLI flag
	for _, arg := range os.Args {
		if arg == "--migrate" || arg == "-m" {
			return true
		}
	}
	// Check env var
	return os.Getenv("RUN_MIGRATIONS") == "true"
}

func ApplyMigrations(db *sql.DB) error {
	m, err := NewMigrateInstance(db)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
