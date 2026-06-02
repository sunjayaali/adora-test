package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	datamocking "adora-test/internal/data/mocking"
	"adora-test/internal/domain"
	"adora-test/internal/service"
	"adora-test/internal/service/mocking"
)

type PollTestSuite struct {
	suite.Suite

	mockCarrier         *mocking.Carrier
	mockEntitlementRepo *mocking.EntitlementRepository
	mockTransactor      *datamocking.Transactor

	poller *service.Poll
}

func (suite *PollTestSuite) SetupTest() {
	suite.mockCarrier = mocking.NewCarrier(suite.T())
	suite.mockEntitlementRepo = mocking.NewEntitlementRepository(suite.T())
	suite.mockTransactor = datamocking.NewTransactor(suite.T())

	suite.poller = service.NewPoll(
		suite.mockCarrier,
		suite.mockEntitlementRepo,
		suite.mockTransactor,
	)
}

func (suite *PollTestSuite) TestPoll() {
	suite.mockCarrier.EXPECT().
		GetStatus(suite.T().Context(), "user1").
		Return(&service.CarrierStatus{
			Status: service.SubscriptionStatusActive,
		}, nil)

	suite.mockTransactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	suite.mockEntitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   userID,
				IsActive: false,
			}, nil
		})

	suite.mockEntitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Equal("user1", entitlement.UserID)
			suite.True(entitlement.IsActive)
			return nil
		})

	err := suite.poller.Poll(suite.T().Context(), "user1")
	suite.NoError(err)
}

func (suite *PollTestSuite) TestPollInactive() {
	suite.mockCarrier.EXPECT().
		GetStatus(suite.T().Context(), "user2").
		Return(&service.CarrierStatus{
			Status: service.SubscriptionStatusInactive,
		}, nil)

	suite.mockTransactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	suite.mockEntitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   userID,
				IsActive: true,
			}, nil
		})

	suite.mockEntitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Equal("user2", entitlement.UserID)
			suite.False(entitlement.IsActive)
			return nil
		})

	err := suite.poller.Poll(suite.T().Context(), "user2")
	suite.NoError(err)
}

func TestPollTestSuite(t *testing.T) {
	suite.Run(t, new(PollTestSuite))
}
