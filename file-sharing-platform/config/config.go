package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

func ConnectDB() (*pgx.Conn, error) {
	var conn *pgx.Conn
	var err error

	dsn := "postgres://postgres:pranav123@postgres:5432/file_sharing_platform"
	retries := 5

	for retries > 0 {
		conn, err = pgx.Connect(context.Background(), dsn)
		if err == nil {
			fmt.Println("Connected to the database successfully!")
			return conn, nil
		}

		fmt.Printf("Failed to connect to database: %v. Retrying in 5 seconds...\n", err)
		time.Sleep(5 * time.Second) // Wait before retrying
		retries--
	}

	return nil, fmt.Errorf("unable to connect to database after retries: %v", err)
}
