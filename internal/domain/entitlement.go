package domain

import "time"

type Source string

const (
	SourceStore       Source = "STORE"
	SourceCarrier     Source = "CARRIER"
	SourceMarketplace Source = "MARKETPLACE"
	SourceNone        Source = "NONE"
)

type Entitlement struct {
	ID              int
	UserID          string
	Source          Source
	IsActive        bool
	ExpiresAt       time.Time
	LastChangedAt   *time.Time
	Reason          string
	LastEventTimeMs int64
	LastPolledAt    *time.Time
}
