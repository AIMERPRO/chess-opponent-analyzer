package domain

import "time"

type RefreshToken struct {
	ID        int64
	UserID    int64
	DeviceID  string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
