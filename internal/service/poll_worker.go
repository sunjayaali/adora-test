package service

import (
	"context"

	"adora-test/internal/data"
	"adora-test/internal/domain"
)

type Poller interface {
	Poll(ctx context.Context, userID string) error
}

type PollWorker struct {
	transactor            data.Transactor
	entitlementRepository EntitlementRepository
	poller                Poller
}

func NewPollWorker(
	transactor data.Transactor,
	entitlementRepository EntitlementRepository,
	poller Poller,
) *PollWorker {
	return &PollWorker{
		transactor:            transactor,
		entitlementRepository: entitlementRepository,
		poller:                poller,
	}
}

func (w *PollWorker) Run(ctx context.Context) error {
	for {
		var entitlements []*domain.Entitlement
		if err := w.transactor.Tx(ctx, func(ctx context.Context) error {
			var err error
			entitlements, err = w.entitlementRepository.ClaimCarrierEntitlements(ctx, 100)
			return err
		}); err != nil {
			return err
		}

		if len(entitlements) == 0 {
			break
		}

		for _, entitlement := range entitlements {
			if err := w.poller.Poll(ctx, entitlement.UserID); err != nil {
				continue
			}
		}
	}

	return nil
}
