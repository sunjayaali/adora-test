package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
	SubscriptionStatusAPIError SubscriptionStatus = "api_error"
)

type CarrierStatus struct {
	Status SubscriptionStatus `json:"status"`
}

type Carrier interface {
	GetStatus(ctx context.Context, userID string) (*CarrierStatus, error)
}

type CarrierClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewCarrierClient(
	httpClient *http.Client,
	baseURL string,
) *CarrierClient {
	return &CarrierClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *CarrierClient) GetStatus(ctx context.Context, userID string) (*CarrierStatus, error) {
	endpoint := fmt.Sprintf(
		"%s/mock/carrier/plan?userId=%s",
		c.baseURL,
		url.QueryEscape(userID),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected status code: %d",
			resp.StatusCode,
		)
	}

	var result CarrierStatus

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	switch result.Status {
	case SubscriptionStatusActive,
		SubscriptionStatusInactive,
		SubscriptionStatusAPIError:
	default:
		return nil, fmt.Errorf(
			"unknown subscription status: %s",
			result.Status,
		)
	}

	return &result, nil
}
