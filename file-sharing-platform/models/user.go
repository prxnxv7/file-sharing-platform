package models

import (
    "time"
)

type User struct {
    ID        int       `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Password  string    `json:"password" db:"password"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
