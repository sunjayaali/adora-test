package service

import (
	"adora-test/internal/data"
	"adora-test/internal/domain"
	"context"
)

type RevokeEntitlementRequest struct {
	UserIDs []string
}

type RevokeEntitlementResponse struct{}

type Revoke struct {
	transactor      data.Transactor
	entitlementRepo EntitlementRepository
}

func NewRevoke(entitlementRepo EntitlementRepository, transactor data.Transactor) *Revoke {
	return &Revoke{
		transactor:      transactor,
		entitlementRepo: entitlementRepo,
	}
}

func (s *Revoke) RevokeEntitlement(ctx context.Context, req *RevokeEntitlementRequest) (*RevokeEntitlementResponse, error) {
	if err := s.transactor.Tx(ctx, func(ctx context.Context) error {
		for _, userID := range req.UserIDs {
			entitlement, err := s.entitlementRepo.FindByUserID(ctx, userID)
			if err != nil {
				return err
			}

			if entitlement == nil {
				continue
			}

			if entitlement.Source != domain.SourceMarketplace {
				continue
			}

			entitlement.IsActive = false
			if err := s.entitlementRepo.Update(ctx, entitlement); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &RevokeEntitlementResponse{}, nil
}
