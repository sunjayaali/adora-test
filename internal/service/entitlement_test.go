package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	"adora-test/internal/domain"
	"adora-test/internal/service"
	"adora-test/internal/service/mocking"
)

type EntitlementServiceTestSuite struct {
	suite.Suite

	entitlementRepo *mocking.EntitlementRepository

	service *service.Entitlement
}

func (suite *EntitlementServiceTestSuite) SetupTest() {
	suite.entitlementRepo = mocking.NewEntitlementRepository(suite.T())

	suite.service = service.NewEntitlementService(suite.entitlementRepo)
}

func (suite *EntitlementServiceTestSuite) TestGetEntitlement_Success() {
	suite.entitlementRepo.EXPECT().
		FindByUserID(context.Background(), "user1").
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				IsActive:      true,
				Source:        "store",
				ExpiresAt:     time.Now().Add(24 * time.Hour),
				LastChangedAt: lo.ToPtr(time.Now()),
				Reason:        "initial grant",
			}, nil
		})

	resp, err := suite.service.GetEntitlement(context.Background(), &service.GetEntitlementRequest{
		UserID: "user1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Entitlement.IsActive)
	suite.Equal("store", string(resp.Entitlement.Source))
}

func (suite *EntitlementServiceTestSuite) TestGetEntitlement_NotFound() {
	suite.entitlementRepo.EXPECT().
		FindByUserID(context.Background(), "user2").
		Return(nil, nil)

	resp, err := suite.service.GetEntitlement(context.Background(), &service.GetEntitlementRequest{
		UserID: "user2",
	})

	suite.Error(err)
	suite.Nil(resp)
	suite.Equal("entitlement not found", err.Error())
}

func TestEntitlementServiceTestSuite(t *testing.T) {
	suite.Run(t, new(EntitlementServiceTestSuite))
}
