package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MockWebhookLogs struct {
	Count   func(context.Context, WebhookLogsListOpts) (int64, error)
	Create  func(context.Context, *types.WebhookLog) error
	GetByID func(context.Context, int64) (*types.WebhookLog, error)
	List    func(context.Context, WebhookLogsListOpts) ([]*types.WebhookLog, int64, error)
}
