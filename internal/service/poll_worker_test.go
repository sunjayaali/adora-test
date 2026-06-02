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

type PollWorkerTestSuite struct {
	suite.Suite

	mockTransactor      *datamocking.Transactor
	mockEntitlementRepo *mocking.EntitlementRepository
	mockPoller          *mocking.Poller

	pollWorker *service.PollWorker
}

func (suite *PollWorkerTestSuite) SetupTest() {
	suite.mockTransactor = datamocking.NewTransactor(suite.T())
	suite.mockEntitlementRepo = mocking.NewEntitlementRepository(suite.T())
	suite.mockPoller = mocking.NewPoller(suite.T())

	suite.pollWorker = service.NewPollWorker(
		suite.mockTransactor,
		suite.mockEntitlementRepo,
		suite.mockPoller,
	)
}

func (suite *PollWorkerTestSuite) TestRun() {
	suite.mockTransactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	entitlementsChan := make(chan []*domain.Entitlement)
	go func() {
		entitlementsChan <- []*domain.Entitlement{
			{UserID: "user1"},
			{UserID: "user2"},
		}
		entitlementsChan <- []*domain.Entitlement{}
		close(entitlementsChan)
	}()

	suite.mockEntitlementRepo.EXPECT().
		ClaimCarrierEntitlements(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, batchSize int) ([]*domain.Entitlement, error) {
			return <-entitlementsChan, nil
		})

	suite.mockPoller.EXPECT().
		Poll(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) error {
			return nil
		})

	err := suite.pollWorker.Run(suite.T().Context())
	suite.NoError(err)
}

func TestPollWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(PollWorkerTestSuite))
}
