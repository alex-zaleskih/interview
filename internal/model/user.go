package model

import "time"

type User struct {
	ID        string
	Email     string
	FullName  string
	Phone     string
	CreatedAt time.Time
}
