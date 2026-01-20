package model

import "time"

type User struct {
	ID        string
	Email     string
	Name      string
	Password  string // hash
	CreatedAt time.Time
}
