package service

import (
	"context"
	"errors"

	"adora-test/internal/domain"
)

type Entitlement struct {
	entitlementRepo EntitlementRepository
}

func NewEntitlementService(entitlementRepo EntitlementRepository) *Entitlement {
	return &Entitlement{entitlementRepo: entitlementRepo}
}

type GetEntitlementRequest struct {
	UserID string
}

type GetEntitlementResponse struct {
	Entitlement *domain.Entitlement
}

func (s *Entitlement) GetEntitlement(ctx context.Context, req *GetEntitlementRequest) (*GetEntitlementResponse, error) {
	entitlement, err := s.entitlementRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if entitlement == nil {
		return nil, errors.New("entitlement not found")
	}

	return &GetEntitlementResponse{
		Entitlement: entitlement,
	}, nil
}
