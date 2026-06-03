package domain

import "time"

type Notification struct {
	ID           int
	Type         string
	UserID       string
	SentAt       *time.Time
	ScheduledFor time.Time
	ExpiresAt    time.Time
}
