package db

import (
	"database/sql"
	"os"
)

func RunMigrations(db *sql.DB) error {
	migrationFiles := []string{
		"migrations/20230822120000_create_products_table.up.sql",
		// Other migrations files can be added if required
	}

	for _, file := range migrationFiles {
		migrationSQL, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			return err
		}
	}

	return nil
}
