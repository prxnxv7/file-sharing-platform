package config

import (
	"context"
	"fmt"
	"os"
	"time"
	"github.com/jackc/pgx/v4"
)

var (
	DbConn   *pgx.Conn
)

func ConnectDB() (*pgx.Conn, error) {
	var conn *pgx.Conn
	var err error

	dsn := fmt.Sprintf("postgres://%s:%s@postgres:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	retries := 5

	for retries > 0 {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err == nil {
			fmt.Println(dsn)

			fmt.Println("Connected to the database successfully!")
			return conn, nil
		}

		fmt.Printf("Failed to connect to database: %v. Retrying in 5 seconds...\n", err)
		time.Sleep(5 * time.Second)
		retries--
	}

	return nil, fmt.Errorf("unable to connect to database after retries: %v", err)
}