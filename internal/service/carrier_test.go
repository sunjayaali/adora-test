package service_test

import (
	"adora-test/internal/service"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CarrierTestSuite struct {
	suite.Suite

	handlerFunc http.HandlerFunc
	testServer  *httptest.Server

	carrier *service.CarrierClient
}

func (suite *CarrierTestSuite) SetupSuite() {
	suite.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.handlerFunc(w, r)
	}))
}

func (suite *CarrierTestSuite) TearDownSuite() {
	suite.testServer.Close()
}

func (suite *CarrierTestSuite) SetupTest() {
	suite.carrier = service.NewCarrierClient(suite.testServer.Client(), suite.testServer.URL)
}

func (suite *CarrierTestSuite) TestGetStatus() {
	suite.handlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "active"}`))
	}

	status, err := suite.carrier.GetStatus(context.Background(), "test-user-id")
	suite.NoError(err)
	suite.Equal(service.SubscriptionStatusActive, status.Status)
}

func (suite *CarrierTestSuite) TestGetStatus_APIError() {
	suite.handlerFunc = func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status": "api_error"}`))
	}

	status, err := suite.carrier.GetStatus(context.Background(), "test-user-id")
	suite.NoError(err)
	suite.Equal(service.SubscriptionStatusAPIError, status.Status)
}

func TestCarrierTestSuite(t *testing.T) {
	suite.Run(t, new(CarrierTestSuite))
}
