package db

import (
	"database/sql"
	"fmt"
	"os"
)

func ApplyMigrations(db *sql.DB,path string) error {
	content, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("error reading migration file: %w", err)
	}
	if _ , err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("error applying migration: %w", err)
	}

	return nil
}