package repository

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/theluckymadman/memotogo/internal/deps"
)

type PGStorage struct {
	DB *sqlx.DB
	*deps.Deps
}

type MigrationCmd int

const (
	UP MigrationCmd = iota
	DOWN
)

func NewPGStorage(dsn string, initMode DBInitMode, migrationsDir string, deps *deps.Deps) (*PGStorage, error) {
	logger := deps.Logger
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("NewStorage: DB open: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("NewStorage: ping DB: %w", err)
	}

	switch initMode {
	case External:
		logger.Info("External managed DB mode is chosen")
	case Internal:
		{
			logger.Info("Internal managed DB mode is chosen")
			if err = runMigration(db, migrationsDir, UP, deps); err != nil {
				return nil, fmt.Errorf("NewStorage: initiate DB: %w", err)
			}
		}
	case ForceInternal:
		{
			logger.Info("Internal with force managed DB mode is chosen")
			if err = runMigration(db, migrationsDir, DOWN, deps); err != nil {
				return nil, fmt.Errorf("NewStorage: initiate DB: %w", err)
			}
			if err = runMigration(db, migrationsDir, UP, deps); err != nil {
				return nil, fmt.Errorf("NewStorage: initiate DB: %w", err)
			}
		}
	default:
		logger.Info("Wrong mode, use external managed DB as default")
	}
	storage := PGStorage{DB: db}
	return &storage, nil
}

func runMigration(db *sqlx.DB, migrationsDir string, cmd MigrationCmd, deps *deps.Deps) error {
	logger := deps.Logger
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("RunMigration: open DB driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("RunMigration: new instance: %w", err)
	}

	switch cmd {
	case UP:
		logger.Info("UP DB")
		if err = m.Up(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				return fmt.Errorf("RunMigration: up DB: %w", err)
			}
		}
	case DOWN:
		logger.Info("DOWN DB")
		if err = m.Down(); err == nil {
			logger.Info("DB down succeeded")
			return nil
		}
		logger.Info("down DB", zap.Error(err))
		logger.Info("trying to down DB with force")
		v, d, err := m.Version()
		logger.Info("Version info", zap.Int("version", int(v)), zap.Bool("dirty", d), zap.Error(err))
		if d || v == 0 {
			logger.Info("Clean dirty DB state flag by force", zap.Int("version", int(v)))
			err = m.Force(int(v))
			if err != nil {
				return fmt.Errorf("RunMigration: down DB: %w", err)
			}
			_, _, err = m.Version()
			if err != nil {
				return fmt.Errorf("RunMigration: down DB: %w", err)
			}
			logger.Info("DB now", zap.Bool("state", d), zap.Int("version", int(v)))
			logger.Info("Try to down again")
			if err = m.Down(); err != nil {
				return fmt.Errorf("RunMigration: down DB: %w", err)
			}
		}
	}
	return nil
}
