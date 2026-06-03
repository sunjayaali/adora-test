package repositories

import (
	"context"
	"time"

	"adora-test/ent"
	"adora-test/ent/notification"
	"adora-test/internal/data"
	"adora-test/internal/domain"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Notification struct {
	manager *data.EntManager
}

func NewNotification(manager *data.EntManager) *Notification {
	return &Notification{manager: manager}
}

func (r *Notification) Insert(ctx context.Context, notification *domain.Notification) error {
	row, err := r.manager.Client(ctx).Notification.Create().
		SetType(notification.Type).
		SetUserID(notification.UserID).
		SetNillableSentAt(notification.SentAt).
		SetScheduledFor(notification.ScheduledFor).
		SetExpiresAt(notification.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}

	notification.ID = row.ID
	return nil
}

func (r *Notification) FindDue(ctx context.Context, now time.Time) ([]*domain.Notification, error) {
	rows, err := r.manager.Client(ctx).Notification.Query().
		Where(
			notification.ScheduledForLTE(now),
			notification.SentAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Map(rows, func(row *ent.Notification, _ int) *domain.Notification {
		return &domain.Notification{
			ID:           row.ID,
			Type:         row.Type,
			UserID:       row.UserID,
			SentAt:       row.SentAt,
			ScheduledFor: row.ScheduledFor,
			ExpiresAt:    row.ExpiresAt,
		}
	}), nil
}

func (r *Notification) Update(ctx context.Context, notification *domain.Notification) error {
	_, err := r.manager.Client(ctx).Notification.UpdateOneID(notification.ID).
		SetNillableSentAt(notification.SentAt).
		SetType(notification.Type).
		SetUserID(notification.UserID).
		SetScheduledFor(notification.ScheduledFor).
		SetExpiresAt(notification.ExpiresAt).
		Save(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
