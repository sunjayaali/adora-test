package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	datamock "adora-test/internal/data/mocking"
	"adora-test/internal/domain"
	"adora-test/internal/service/mocking"
)

type WebhookStoreServiceTestSuite struct {
	suite.Suite

	storeEventRepo  *mocking.StoreEventRepository
	entitlementRepo *mocking.EntitlementRepository
	transactor      *datamock.Transactor

	service *WebhookStoreService
}

func (suite *WebhookStoreServiceTestSuite) SetupTest() {
	suite.storeEventRepo = mocking.NewStoreEventRepository(suite.T())
	suite.entitlementRepo = mocking.NewEntitlementRepository(suite.T())
	suite.transactor = datamock.NewTransactor(suite.T())

	suite.service = NewWebhookStoreService(
		suite.storeEventRepo,
		suite.entitlementRepo,
		suite.transactor,
	)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_NewEvent() {
	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			suite.Require().Equal("event1", eventID)

			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			suite.Require().Equal("event1", storeEvent.EventID)
			suite.Require().Equal("user1", storeEvent.UserID)
			suite.Require().Equal(domain.StoreEventTypeInitialPurchase, storeEvent.Type)
			suite.Require().Equal(int64(1), storeEvent.EventTimeMs)
			suite.Require().Equal("product1", storeEvent.ProductID)

			return nil
		})
	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			suite.Require().Equal("user1", userID)

			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			suite.Require().Equal("event1", storeEvent.EventID)
			suite.Require().Equal("user1", storeEvent.UserID)
			suite.Require().Equal(domain.StoreEventTypeInitialPurchase, storeEvent.Type)
			suite.Require().Equal(int64(1), storeEvent.EventTimeMs)
			suite.Require().Equal("product1", storeEvent.ProductID)

			return nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			suite.Require().Equal("user1", userID)

			return nil, nil
		})
	suite.entitlementRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("INITIAL_PURCHASE", entitlement.Reason)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event1",
		UserID:      "user1",
		Type:        "INITIAL_PURCHASE",
		EventTimeMs: 1,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_DuplicateEvent() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			suite.Require().Equal("event1", eventID)

			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event1",
		UserID:      "user1",
		Type:        "INITIAL_PURCHASE",
		EventTimeMs: 1,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_NotDuplicateEvent() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 2,
				ProductID:   "product1",
			}, nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "INITIAL_PURCHASE",
		EventTimeMs: 1,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_EntitlementInsert() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return nil, nil
		})
	suite.entitlementRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("INITIAL_PURCHASE", entitlement.Reason)
			suite.Require().True(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event1",
		UserID:      "user1",
		Type:        "INITIAL_PURCHASE",
		EventTimeMs: 1,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_EntitlementUpdate() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "INITIAL_PURCHASE",
				IsActive: true,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("CANCELLATION", entitlement.Reason)
			suite.Require().False(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "CANCELLATION",
		EventTimeMs: 2,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_EntitlementUpdate_Cancellation() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "INITIAL_PURCHASE",
				IsActive: true,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("CANCELLATION", entitlement.Reason)
			suite.Require().False(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "CANCELLATION",
		EventTimeMs: 2,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_EntitlementUpdate_Change() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "CANCELLATION",
				IsActive: true,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("INITIAL_PURCHASE", entitlement.Reason)
			suite.Require().True(entitlement.IsActive)
			suite.Require().Equal(time.UnixMilli(2).Add(30*24*time.Hour), entitlement.ExpiresAt)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "INITIAL_PURCHASE",
		EventTimeMs: 2,
		ProductID:   "product1",
	})

	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_ToCancellation() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "INITIAL_PURCHASE",
				IsActive: true,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("CANCELLATION", entitlement.Reason)
			suite.Require().False(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "CANCELLATION",
		EventTimeMs: 2,
		ProductID:   "product1",
	})
	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_ToRenewal() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "CANCELLATION",
				IsActive: false,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("RENEWAL", entitlement.Reason)
			suite.Require().True(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "RENEWAL",
		EventTimeMs: 2,
		ProductID:   "product1",
	})
	suite.NoError(err)
	suite.NotNil(resp)
}

func (suite *WebhookStoreServiceTestSuite) TestIngest_ToUncancellation() {
	suite.storeEventRepo.EXPECT().
		FindByEventID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, eventID string) (*domain.StoreEvent, error) {
			return nil, nil
		})
	suite.storeEventRepo.EXPECT().
		Insert(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, storeEvent *domain.StoreEvent) error {
			return nil
		})

	suite.transactor.EXPECT().
		Tx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, f func(context.Context) error) error {
			return f(ctx)
		})

	suite.storeEventRepo.EXPECT().
		FindLatestByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.StoreEvent, error) {
			return &domain.StoreEvent{
				EventID:     "event1",
				UserID:      "user1",
				Type:        domain.StoreEventTypeInitialPurchase,
				EventTimeMs: 1,
				ProductID:   "product1",
			}, nil
		})

	suite.entitlementRepo.EXPECT().
		FindByUserID(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, userID string) (*domain.Entitlement, error) {
			return &domain.Entitlement{
				UserID:   "user1",
				Source:   domain.Source(EntitlementSourceStore),
				Reason:   "CANCELLATION",
				IsActive: false,
			}, nil
		})
	suite.entitlementRepo.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, entitlement *domain.Entitlement) error {
			suite.Require().Equal("user1", entitlement.UserID)
			suite.Require().Equal(domain.Source(EntitlementSourceStore), entitlement.Source)
			suite.Require().Equal("UN_CANCELLATION", entitlement.Reason)
			suite.Require().True(entitlement.IsActive)

			return nil
		})

	resp, err := suite.service.Ingest(context.Background(), &IngestStoreWebhookRequest{
		EventID:     "event2",
		UserID:      "user1",
		Type:        "UN_CANCELLATION",
		EventTimeMs: 2,
		ProductID:   "product1",
	})
	suite.NoError(err)
	suite.NotNil(resp)
}

func Test(t *testing.T) {
	suite.Run(t, new(WebhookStoreServiceTestSuite))
}
