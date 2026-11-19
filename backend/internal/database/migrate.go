package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunMigrations(
	ctx context.Context,
	db *sql.DB,
	directory string,
) error {
	// Создаём таблицу учёта миграций.
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Читаем файлы миграций.
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(strings.ToLower(name), ".sql") {
			continue
		}

		path := filepath.Join(directory, name)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		// Начинаем транзакцию.
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}

		// Проверяем, выполнялась ли миграция раньше.
		var applied bool

		err = tx.QueryRowContext(
			ctx,
			`SELECT EXISTS (
				SELECT 1
				FROM schema_migrations
				WHERE version = $1
			)`,
			name,
		).Scan(&applied)

		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		// Если миграция уже выполнена, пропускаем её.
		if applied {
			_ = tx.Rollback()
			continue
		}

		// Выполняем SQL-код миграции.
		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", name, err)
		}

		// Сохраняем информацию о выполненной миграции.
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`,
			name,
		)

		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("save migration %s: %w", name, err)
		}

		// Подтверждаем транзакцию.
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}

		fmt.Printf("[MIGRATION] Applied: %s\n", name)
	}

	return nil
}