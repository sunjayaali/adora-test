package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"adora-test/internal/domain"
	"adora-test/internal/service"
	"adora-test/internal/service/mocking"
)

type NotificationTestSuite struct {
	suite.Suite

	mockEntitlementRepo  *mocking.EntitlementRepository
	mockNotificationRepo *mocking.NotificationRepository

	service *service.Notification
}

func (suite *NotificationTestSuite) SetupTest() {
	suite.mockEntitlementRepo = mocking.NewEntitlementRepository(suite.T())
	suite.mockNotificationRepo = mocking.NewNotificationRepository(suite.T())
	suite.service = service.NewNotification(suite.mockEntitlementRepo, suite.mockNotificationRepo)
}

func (suite *NotificationTestSuite) TestScheduleExpiringNotifications() {
	suite.mockEntitlementRepo.EXPECT().
		FindExpiring(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, before time.Time) ([]*domain.Entitlement, error) {
			return []*domain.Entitlement{
				{UserID: "user1", ExpiresAt: time.Now().Add(23 * time.Hour)},
			}, nil
		})

	suite.mockNotificationRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, notification *domain.Notification) error {
			suite.Equal("PREMIUM_EXPIRES_SOON", notification.Type)
			return nil
		})

	err := suite.service.ScheduleExpiringNotifications(suite.T().Context())
	suite.NoError(err)
}

func (suite *NotificationTestSuite) TestSendDueNotifications() {
	suite.mockNotificationRepo.EXPECT().
		FindDue(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, now time.Time) ([]*domain.Notification, error) {
			return []*domain.Notification{
				{ID: 1, UserID: "user1", Type: "PREMIUM_EXPIRES_SOON"},
			}, nil
		})

	suite.mockNotificationRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, notification *domain.Notification) error {
			suite.Equal(1, notification.ID)
			suite.NotNil(notification.SentAt)
			return nil
		})

	err := suite.service.SendDueNotifications(suite.T().Context())
	suite.NoError(err)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}
