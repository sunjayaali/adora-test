package domain

import "time"

type StoreEventType string

const (
	StoreEventTypeInitialPurchase StoreEventType = "INITIAL_PURCHASE"
	StoreEventTypeRenewal         StoreEventType = "RENEWAL"
	StoreEventTypeCancellation    StoreEventType = "CANCELLATION"
	StoreEventTypeBillingIssue    StoreEventType = "BILLING_ISSUE"
	StoreEventTypeExpiration      StoreEventType = "EXPIRATION"
	StoreEventTypeUnCancellation  StoreEventType = "UN_CANCELLATION"
)

type StoreEvent struct {
	ID          int
	EventID     string
	UserID      string
	Type        StoreEventType
	EventTimeMs int64
	ProductID   string
	ReceivedAt  time.Time
}
