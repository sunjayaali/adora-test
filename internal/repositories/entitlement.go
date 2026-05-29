package repositories

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"adora-test/ent"
	"adora-test/ent/entitlement"
	"adora-test/internal/data"
	"adora-test/internal/domain"
)

type Entitlement struct {
	manager *data.EntManager
}

func NewEntitlement(manager *data.EntManager) *Entitlement {
	return &Entitlement{manager: manager}
}

func (e Entitlement) FindByUserID(ctx context.Context, userID string) (*domain.Entitlement, error) {
	row, err := e.manager.Client(ctx).Entitlement.Query().
		Where(entitlement.UserID(userID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	return e.toDomain(row), nil
}

func (e Entitlement) Insert(ctx context.Context, entitlement *domain.Entitlement) error {
	row, err := e.manager.Client(ctx).Entitlement.Create().
		SetUserID(entitlement.UserID).
		SetSource(string(entitlement.Source)).
		SetIsActive(entitlement.IsActive).
		SetExpiresAt(entitlement.ExpiresAt).
		SetNillableLastChangedAt(entitlement.LastChangedAt).
		SetReason(entitlement.Reason).
		Save(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	entitlement.ID = row.ID

	return nil
}

func (e Entitlement) Update(ctx context.Context, entitlement *domain.Entitlement) error {
	_, err := e.manager.Client(ctx).Entitlement.UpdateOneID(entitlement.ID).
		SetIsActive(entitlement.IsActive).
		SetExpiresAt(entitlement.ExpiresAt).
		SetLastChangedAt(time.Now().UTC()).
		SetReason(entitlement.Reason).
		Save(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (e Entitlement) toDomain(row *ent.Entitlement) *domain.Entitlement {
	return &domain.Entitlement{
		ID:            row.ID,
		UserID:        row.UserID,
		Source:        domain.Source(row.Source),
		IsActive:      row.IsActive,
		ExpiresAt:     row.ExpiresAt,
		LastChangedAt: row.LastChangedAt,
		Reason:        row.Reason,
	}
}
