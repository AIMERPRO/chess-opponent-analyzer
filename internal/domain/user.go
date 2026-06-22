package domain

import "time"

type User struct {
	ID              int64
	Username        string
	LichessUsername string
	Password        string
	CreatedAt       time.Time
}
