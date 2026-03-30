package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB wraps GORM and the underlying sql.DB for shared use across all stores.
type DB struct {
	Gorm *gorm.DB
	SQL  *sql.DB
}

// Connect opens the database, verifies connectivity, and runs goose migrations.
func Connect(ctx context.Context, dsn string) (*DB, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db: %w", err)
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := runMigrations(ctx, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return &DB{Gorm: gdb, SQL: sqlDB}, nil
}

// Close shuts down the shared database connection pool.
func (db *DB) Close() {
	if db.SQL != nil {
		_ = db.SQL.Close()
	}
}
