package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	datamock "adora-test/internal/data/mocking"
	"adora-test/internal/domain"
	"adora-test/internal/service"
	"adora-test/internal/service/mocking"
)

type RevokeTestSuite struct {
	suite.Suite

	entitlementRepo *mocking.EntitlementRepository
	transactor      *datamock.Transactor

	service *service.Revoke
}

func (suite *RevokeTestSuite) SetupTest() {
	suite.entitlementRepo = mocking.NewEntitlementRepository(suite.T())
	suite.transactor = datamock.NewTransactor(suite.T())
	suite.service = service.NewRevoke(suite.entitlementRepo, suite.transactor)
}

func (suite *RevokeTestSuite) TestRevokeEntitlement_Success() {
	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   userID,
				IsActive: true,
				Source:   domain.SourceMarketplace,
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.False(entitlement.IsActive)
			return nil
		})

	req := &service.RevokeEntitlementRequest{
		UserIDs: []string{"user1"},
	}

	resp, err := suite.service.RevokeEntitlement(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *RevokeTestSuite) TestRevokeEntitlement_UserNotFound() {
	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		Return(nil, nil)

	req := &service.RevokeEntitlementRequest{
		UserIDs: []string{"user2"},
	}

	resp, err := suite.service.RevokeEntitlement(context.Background(), req)
	suite.NoError(err)
	suite.NotNil(resp)
}

func TestRevokeTestSuite(t *testing.T) {
	suite.Run(t, new(RevokeTestSuite))
}
