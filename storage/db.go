package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var (
	// dsn = "postgres://postgres:123456@localhost/postgres?sslmode=disable"

	dbHost = "DB_HOST"
	dbPort = "DB_PORT"
	dbUser = "DB_USER"
	dbPass = "DB_PASSWORD"
	dbName = "DB_NAME"

	driver  = "postgres"
	timeout = 5 * time.Second
)

// Open opens a database with the current configured driver and data source
// name.
//
// The database will be probed a simple constant-wait retry strategy; if opening
// the database fails within the specified timeout duration an error is
// returned.
func Open() (*sql.DB, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv(dbHost),
		os.Getenv(dbPort),
		os.Getenv(dbUser),
		os.Getenv(dbPass),
		os.Getenv(dbName),
	)

	db, err := sql.Open(driver, connString)
	if err != nil {
		// DSN parsing or initialization error
		return nil, err
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(4)

	timer := time.After(timeout)
	for i := 0; ; i++ {
		err := db.Ping()
		if err == nil {
			break
		}

		log.Printf("db: %v retrying (attempt: %d)", err, i)
		time.Sleep(1 * time.Second)
		select {
		case <-timer:
			return nil, err
		default:
		}
	}
	return db, nil
}
