package repositories

import (
	"context"

	"github.com/pkg/errors"

	"adora-test/ent"
	"adora-test/ent/storeevent"
	"adora-test/internal/data"
	"adora-test/internal/domain"
	"adora-test/internal/service"
)

var _ service.StoreEventRepository = (*StoreEventRepository)(nil)

type StoreEventRepository struct {
	manager *data.EntManager
}

func NewStoreEventRepository(manager *data.EntManager) *StoreEventRepository {
	return &StoreEventRepository{manager: manager}
}

func (r StoreEventRepository) Insert(ctx context.Context, storeEvent *domain.StoreEvent) error {
	row, err := r.manager.Client(ctx).StoreEvent.Create().
		SetEventID(storeEvent.EventID).
		SetUserID(storeEvent.UserID).
		SetType(string(storeEvent.Type)).
		SetEventTimeMs(storeEvent.EventTimeMs).
		SetProductID(storeEvent.ProductID).
		SetReceivedAt(storeEvent.ReceivedAt).
		Save(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	storeEvent.ID = row.ID

	return nil
}

func (r StoreEventRepository) FindByEventID(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
	row, err := r.manager.Client(ctx).StoreEvent.Query().
		Where(storeevent.EventIDEQ(eventID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	return &domain.StoreEvent{
		EventID:     row.EventID,
		UserID:      row.UserID,
		Type:        domain.StoreEventType(row.Type),
		EventTimeMs: row.EventTimeMs,
		ProductID:   row.ProductID,
		ReceivedAt:  row.ReceivedAt,
	}, nil
}

func (r StoreEventRepository) FindLatestByUserID(ctx context.Context, userID string) (*domain.StoreEvent, error) {
	row, err := r.manager.Client(ctx).StoreEvent.Query().
		Where(storeevent.UserID(userID)).
		Order(ent.Desc(storeevent.FieldEventTimeMs)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	return r.toDomain(row), nil
}

func (r StoreEventRepository) toDomain(row *ent.StoreEvent) *domain.StoreEvent {
	return &domain.StoreEvent{
		EventID:     row.EventID,
		UserID:      row.UserID,
		Type:        domain.StoreEventType(row.Type),
		EventTimeMs: row.EventTimeMs,
		ProductID:   row.ProductID,
		ReceivedAt:  row.ReceivedAt,
	}
}
