package logging

import (
	"context"
	"fmt"

	"adora-test/internal/service"
)

type Poll struct {
	poll service.Poller
}

func NewPoll(poll service.Poller) *Poll {
	return &Poll{poll: poll}
}

func (p *Poll) Poll(ctx context.Context, userID string) error {
	fmt.Println("Polling user:", userID)
	if err := p.poll.Poll(ctx, userID); err != nil {
		fmt.Println("Error polling user:", err)

		return err
	}

	return nil
}
