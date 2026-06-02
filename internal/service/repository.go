package service

import (
	"adora-test/internal/domain"
	"context"
)

type StoreEventRepository interface {
	FindByEventID(ctx context.Context, eventID string) (*domain.StoreEvent, error)
	Insert(ctx context.Context, storeEvent *domain.StoreEvent) error
	FindLatestByUserID(ctx context.Context, userID string) (*domain.StoreEvent, error)
}

type EntitlementRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Entitlement, error)
	Insert(ctx context.Context, entitlement *domain.Entitlement) error
	Update(ctx context.Context, entitlement *domain.Entitlement) error
	ClaimCarrierEntitlements(ctx context.Context, batchSize int) ([]*domain.Entitlement, error)
}
