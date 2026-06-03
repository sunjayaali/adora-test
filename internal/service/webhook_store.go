package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"adora-test/internal/data"
	"adora-test/internal/domain"
)

type WebhookStoreService struct {
	storeEventRepository  StoreEventRepository
	entitlementRepository EntitlementRepository

	transactor data.Transactor
}

const (
	EntitlementSourceStore = "STORE"
)

var ErrInvalidWebhookRequest = errors.New("invalid webhook request")

type IngestStoreWebhookRequest struct {
	EventID     string
	UserID      string
	Type        string
	EventTimeMs int64
	ProductID   string
}

type IngestStoreWebhookResponse struct{}

func NewWebhookStoreService(repo StoreEventRepository, entitlementRepo EntitlementRepository, transactor data.Transactor) *WebhookStoreService {
	return &WebhookStoreService{storeEventRepository: repo, entitlementRepository: entitlementRepo, transactor: transactor}
}

func (s WebhookStoreService) Ingest(ctx context.Context, req *IngestStoreWebhookRequest) (*IngestStoreWebhookResponse, error) {
	storeEvent, err := s.storeEventRepository.FindByEventID(ctx, req.EventID)
	if err != nil {
		return nil, err
	}

	if storeEvent != nil {
		return &IngestStoreWebhookResponse{}, nil
	}

	var response *IngestStoreWebhookResponse
	if err := s.transactor.Tx(ctx, func(ctx context.Context) error {
		storeEventType := domain.StoreEventType(req.Type)
		newStoreEvent := &domain.StoreEvent{
			EventID:     strings.TrimSpace(req.EventID),
			UserID:      strings.TrimSpace(req.UserID),
			Type:        storeEventType,
			EventTimeMs: req.EventTimeMs,
			ProductID:   strings.TrimSpace(req.ProductID),
			ReceivedAt:  time.Now().UTC(),
		}
		if err = s.storeEventRepository.Insert(ctx, newStoreEvent); err != nil {
			return err
		}

		latestStoreEvent, err := s.storeEventRepository.FindLatestByUserID(ctx, req.UserID)
		if err != nil {
			return err
		}
		if req.EventTimeMs < latestStoreEvent.EventTimeMs {
			response = &IngestStoreWebhookResponse{}
			return nil
		}

		entitlement := &domain.Entitlement{
			UserID: req.UserID,
			Source: domain.Source(EntitlementSourceStore),
			Reason: req.Type,
		}
		switch domain.StoreEventType(req.Type) {
		case domain.StoreEventTypeInitialPurchase:
			entitlement.IsActive = true
			entitlement.ExpiresAt = time.UnixMilli(req.EventTimeMs).Add(30 * 24 * time.Hour)
		case domain.StoreEventTypeRenewal:
			entitlement.IsActive = true
			entitlement.ExpiresAt = time.UnixMilli(req.EventTimeMs).Add(30 * 24 * time.Hour)
		case domain.StoreEventTypeUnCancellation:
			entitlement.IsActive = true
			entitlement.ExpiresAt = time.UnixMilli(req.EventTimeMs).Add(30 * 24 * time.Hour)
		case domain.StoreEventTypeCancellation, domain.StoreEventTypeBillingIssue, domain.StoreEventTypeExpiration:
			entitlement.IsActive = false
		}

		currentEntitlement, err := s.entitlementRepository.FindByUserID(ctx, req.UserID)
		if err != nil {
			return err
		}

		if currentEntitlement == nil {
			if err := s.entitlementRepository.Insert(ctx, entitlement); err != nil {
				return err
			}
		} else {
			entitlement.ID = currentEntitlement.ID
			if err := s.entitlementRepository.Update(ctx, entitlement); err != nil {
				return err
			}
		}

		response = &IngestStoreWebhookResponse{}

		return nil
	}); err != nil {
		return nil, err
	}

	return response, nil
}
