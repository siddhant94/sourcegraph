package types

import (
	"net/http"
	"time"
)

type WebhookLog struct {
	ID                int64
	ReceivedAt        time.Time
	ExternalServiceID *int64
	Request           WebhookLogRequest
	Error             *string
}

type WebhookLogRequest struct {
	Headers http.Header
	Body    []byte
}
