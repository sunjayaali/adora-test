package data

import (
	"context"

	"github.com/pkg/errors"

	"adora-test/ent"
)

var _ Transactor = (*EntManager)(nil)

type Transactor interface {
	Tx(ctx context.Context, f func(context.Context) error) error
}

type EntManager struct {
	client *ent.Client
}

func NewEntManager(client *ent.Client) *EntManager {
	return &EntManager{client: client}
}

func (m *EntManager) Tx(ctx context.Context, f func(context.Context) error) error {
	if tx := ent.TxFromContext(ctx); tx != nil {
		return f(ctx)
	}

	tx, err := m.client.Tx(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := f(ent.NewTxContext(ctx, tx)); err != nil {
		_ = tx.Rollback()

		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *EntManager) Client(ctx context.Context) *ent.Client {
	if tx := ent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}

	return m.client
}
