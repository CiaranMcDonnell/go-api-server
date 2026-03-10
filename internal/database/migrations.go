package database

import (
	"database/sql"
	"fmt"

	"github.com/ciaranmcdonnell/go-api-server/internal/database/migrations"
	"github.com/ciaranmcdonnell/go-api-server/pkg/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations() (*sql.DB, error) {
	config, err := utils.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	db, err := sql.Open("pgx", config.DBSource)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	if err := migrations.ApplyMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("error applying migrations: %w", err)
	}

	return db, nil
}
