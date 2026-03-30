package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func runMigrations(ctx context.Context, sqlDB *sql.DB) error {
	goose.SetBaseFS(migrationFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := baselineLegacySchema(ctx, sqlDB); err != nil {
		return fmt.Errorf("baseline legacy schema: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func baselineLegacySchema(ctx context.Context, sqlDB *sql.DB) error {
	version, err := goose.EnsureDBVersionContext(ctx, sqlDB)
	if err != nil {
		return err
	}
	if version != 0 {
		return nil
	}

	exists, err := legacySchemaExists(ctx, sqlDB)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	_, err = sqlDB.ExecContext(ctx, `
		INSERT INTO goose_db_version (version_id, is_applied)
		SELECT 1, TRUE
		WHERE NOT EXISTS (
			SELECT 1
			FROM goose_db_version
			WHERE version_id = 1 AND is_applied = TRUE
		)
	`)
	return err
}

func legacySchemaExists(ctx context.Context, sqlDB *sql.DB) (bool, error) {
	var exists bool
	err := sqlDB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
