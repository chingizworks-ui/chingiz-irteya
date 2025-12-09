package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ApplyMigrations(ctx context.Context, connString, migrationsPath string) error {
	sqlData, err := os.ReadFile(migrationsPath)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	migrationCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(migrationCtx, connString)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	queries := strings.Split(string(sqlData), ";")
	for _, raw := range queries {
		stmt := strings.TrimSpace(raw)
		if stmt == "" {
			continue
		}
		if _, errExec := pool.Exec(migrationCtx, stmt); errExec != nil {
			return fmt.Errorf("execute migration statement: %w", errExec)
		}
	}
	return nil
}
