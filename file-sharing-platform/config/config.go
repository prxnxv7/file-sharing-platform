package config

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v4"
)

func ConnectDB() (*pgx.Conn, error) {
    conn, err := pgx.Connect(context.Background(), "postgres://postgres:pranav123@localhost:5432/file_sharing_platform")
    if err != nil {
        return nil, fmt.Errorf("unable to connect to database: %v", err)
    }
    return conn, nil
}
