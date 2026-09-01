package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const migrationPath = "file://migrations"
const undefinedTableCode = "42P01"
const foreignKeyViolationCode = "23503"

func openMigrator() (*migrate.Migrate, *sql.DB, error) {
	connString := os.Getenv("KUBEGUARD_POSTGRES_URL")
	if connString == "" {
		return nil, nil, fmt.Errorf("KUBEGUARD_POSTGRES_URL environment is not set")
	}

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db: %w", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres", driver)
	if err != nil {
		sqlDB.Close()
		return nil, nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, sqlDB, nil
}

func MigrateUp() error {
	m, sqlDB, err := openMigrator()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	if err := m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func MigrateDown(steps int) error {
	m, sqlDB, err := openMigrator()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return m.Steps(-steps)
}

func MigrationStatus() (version uint, dirty bool, err error) {
	m, sqlDB, err := openMigrator()
	if err != nil {
		return 0, false, err
	}
	defer sqlDB.Close()

	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

func GenerateUUID() uuid.UUID {
	myuuid := uuid.New()
	return myuuid
}

func IsSchemaNotExist(err error) bool {
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) {
		return pqErr.Code == undefinedTableCode
	}
	return false
}

func IsForeignKeyViolation(err error) bool {
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) {
		return pqErr.Code == foreignKeyViolationCode
	}
	return false
}
