package service

import (
	"context"
	"fmt"

	"adora-test/internal/data"
)

type Poll struct {
	carrier         Carrier
	entitlementRepo EntitlementRepository
	transactor      data.Transactor
}

func NewPoll(carrier Carrier, entitlementRepo EntitlementRepository, transactor data.Transactor) *Poll {
	return &Poll{
		carrier:         carrier,
		entitlementRepo: entitlementRepo,
		transactor:      transactor,
	}
}

func (p *Poll) Poll(ctx context.Context, userID string) error {
	status, err := p.carrier.GetStatus(ctx, userID)
	if err != nil {
		return err
	}

	if err := p.transactor.Tx(ctx, func(ctx context.Context) error {
		entitlement, err := p.entitlementRepo.FindByUserID(ctx, userID)
		if err != nil {
			return err
		}

		switch status.Status {
		case SubscriptionStatusActive:
			entitlement.IsActive = true
		case SubscriptionStatusInactive:
			entitlement.IsActive = false
		case SubscriptionStatusAPIError:
			return fmt.Errorf("API error for user %s", userID)
		}

		if err := p.entitlementRepo.Update(ctx, entitlement); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
