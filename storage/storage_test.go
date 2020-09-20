package storage

import (
	"database/sql"
	"testing"
)

func setupDB(db *sql.DB) error {
	queries := []string{
		`DROP SCHEMA IF EXISTS test CASCADE`,
		`CREATE SCHEMA test`,
		`CREATE TABLE test.product (
            id SERIAL PRIMARY KEY,
            sku text UNIQUE NOT NULL,
            attrs JSONB
        )`,
		`SET search_path TO test,public`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func TestSetup(t *testing.T) {
	db1, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	db2, err := Open()
	if err != nil {
		t.Fatal(err)
	}

	err = setupDB(db1)
	if err != nil {
		t.Fatal(err)
	}
	err = setupDB(db2)
	if err != nil {
		t.Fatal(err)
	}
}
