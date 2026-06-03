package service

import (
	"context"
	"fmt"
	"time"

	"adora-test/internal/domain"

	"github.com/samber/lo"
)

type Notification struct {
	entitlementRepo  EntitlementRepository
	notificationRepo NotificationRepository
}

func NewNotification(entitlementRepo EntitlementRepository, notificationRepo NotificationRepository) *Notification {
	return &Notification{
		entitlementRepo:  entitlementRepo,
		notificationRepo: notificationRepo,
	}
}

func (n *Notification) ScheduleExpiringNotifications(ctx context.Context) error {
	expiringEntitlements, err := n.entitlementRepo.FindExpiring(ctx, time.Now().Add(24*time.Hour))
	if err != nil {
		return err
	}

	for _, entitlement := range expiringEntitlements {
		notification := &domain.Notification{
			Type:         "PREMIUM_EXPIRES_SOON",
			UserID:       entitlement.UserID,
			ScheduledFor: entitlement.ExpiresAt.Add(-24 * time.Hour),
			ExpiresAt:    entitlement.ExpiresAt,
		}
		if err := n.notificationRepo.Insert(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}

func (n *Notification) SendDueNotifications(ctx context.Context) error {
	dueNotifications, err := n.notificationRepo.FindDue(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, notification := range dueNotifications {
		// Here you would implement the logic to send the notification to the user.
		// For example, you could send an email or push notification.
		// This is a placeholder for the actual sending logic.
		fmt.Println("Sending notification to user:", notification.UserID, "Type:", notification.Type)

		notification.SentAt = lo.ToPtr(time.Now())
		if err := n.notificationRepo.Update(ctx, notification); err != nil {
			return err
		}
	}

	return nil
}
