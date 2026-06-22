package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectWithRetry(dsn string, maxRetries int) (*sql.DB, error) {
	var db *sql.DB
	var err error
	backoff := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to the database (attempt %d/%d)", i+1, maxRetries)
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Println("Successfully connected to the database")
				db.SetMaxOpenConns(25) // Set the maximum number of open connections
				db.SetMaxIdleConns(25) // Set the maximum number of idle connections
				db.SetConnMaxLifetime(5 * time.Minute) // Set the maximum lifetime of a connection
				
				return db, nil
			}
		}
		log.Printf("Failed to connect to the database: %v", err)
		time.Sleep(backoff)
		backoff *= 2 // Exponential backoff
	}


	return nil, err
}
