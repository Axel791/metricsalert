package db

import (
	"context"
	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

// ConnectDB - подключение к базе данных, применение миграций
func ConnectDB(databaseDSN string, cfg *config.Config) (*sqlx.DB, error) {
	if databaseDSN != "" {
		db, err := sqlx.Connect("postgres", databaseDSN)
		if err != nil {
			return nil, err
		}

		err = appleMigration(db, cfg)
		if err != nil {
			_ = db.Close()
			return nil, err
		}

		return db, nil
	}
	return nil, nil
}

// appleMigration - Применение миграций
func appleMigration(dbConn *sqlx.DB, cfg *config.Config) error {
	if err := goose.RunContext(context.Background(), "up", dbConn.DB, cfg.MigrationsPath); err != nil {
		return err
	}
	return nil
}
