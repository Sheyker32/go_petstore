package postgres

import (
	"swagger_petstore/config"
	"swagger_petstore/internal/repository"

	"github.com/golang-migrate/migrate"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewPostgresDB(conf *config.DBConfig, logger *zap.Logger) *sqlx.DB {
	db, err := sqlx.Open("postgres", conf.GetDBURL())
	if err != nil {
		logger.Fatal("Failed to connect to database: ", zap.Error(err))
	}

	m := repository.NewMigration(conf)
	if err := m.Up(); err != nil && err.Error() != migrate.ErrNoChange.Error() {
		logger.Fatal("error migrate: ", zap.Error(err))
	}
	return db
}
