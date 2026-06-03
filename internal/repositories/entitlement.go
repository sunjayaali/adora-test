package repositories

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/pkg/errors"

	"adora-test/ent"
	"adora-test/ent/entitlement"
	"adora-test/ent/notification"
	"adora-test/internal/data"
	"adora-test/internal/domain"
	"adora-test/internal/service"
)

var _ service.EntitlementRepository = (*Entitlement)(nil)

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

func (e Entitlement) ClaimCarrierEntitlements(ctx context.Context, limit int) ([]*domain.Entitlement, error) {
	client := e.manager.Client(ctx)

	rows, err := client.Entitlement.
		Query().
		Where(
			entitlement.SourceEQ(string(domain.SourceCarrier)),
			entitlement.Or(
				entitlement.LastPolledAtIsNil(),
				entitlement.LastPolledAtLT(time.Now().Add(-5*time.Minute)),
			),
		).
		ForUpdate(sql.WithLockAction(sql.SkipLocked)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(rows) == 0 {
		return []*domain.Entitlement{}, nil
	}

	now := time.Now().UTC()

	ids := make([]int, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}

	if _, err = client.Entitlement.
		Update().
		Where(entitlement.IDIn(ids...)).
		SetLastPolledAt(now).
		Save(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]*domain.Entitlement, 0, len(rows))
	for _, r := range rows {
		d := e.toDomain(r)
		d.LastPolledAt = &now
		result = append(result, d)
	}

	return result, nil
}

func (r Entitlement) FindExpiring(ctx context.Context, before time.Time) ([]*domain.Entitlement, error) {
	rows, err := r.manager.Client(ctx).Entitlement.
		Query().
		Where(func(s *sql.Selector) {
			n := sql.Table(notification.Table)

			s.LeftJoin(n).
				On(s.C(entitlement.FieldUserID), n.C(notification.FieldUserID)).
				Where(
					sql.And(
						sql.EQ(s.C(entitlement.FieldIsActive), true),
						sql.NotNull(s.C(entitlement.FieldExpiresAt)),
						sql.LTE(s.C(entitlement.FieldExpiresAt), before),
						sql.Or(
							sql.IsNull(n.C(notification.FieldID)),
							sql.And(sql.EQ(n.C(notification.FieldType), "PREMIUM_EXPIRES_SOON"),
								sql.Not(
									sql.ColumnsEQ(s.C(entitlement.FieldExpiresAt), n.C(notification.FieldExpiresAt)),
								),
							),
						),
					),
				)
		}).
		All(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]*domain.Entitlement, 0, len(rows))
	for _, row := range rows {
		d := r.toDomain(row)
		result = append(result, d)
	}

	return result, nil
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
		LastPolledAt:  row.LastPolledAt,
	}
}
