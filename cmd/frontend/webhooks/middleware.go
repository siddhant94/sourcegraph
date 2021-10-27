package webhooks

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type contextKey string

var setterContextKey = contextKey("webhook setter")

type contextFunc func(int64)

// SetExternalServiceID attaches a specific external service ID to the current
// webhook request for logging purposes.
func SetExternalServiceID(ctx context.Context, id int64) {
	if setter, ok := ctx.Value(setterContextKey).(contextFunc); ok {
		setter(id)
	} else {
		log15.Error("cannot get setter from context; this likely means that SetExternalServiceID has been called from outside a HTTP handler wrapped in the WebhookLogger middleware")
	}
}

type LogMiddleware struct {
	store *database.WebhookLogsStore
}

func NewLogMiddleware(store *database.WebhookLogsStore) *LogMiddleware {
	return &LogMiddleware{store}
}

func (mw *LogMiddleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: refactor into a utility function.
		shouldLog := true
		if logging := conf.Get().WebhookLogging; logging != nil && logging.Enabled != nil {
			shouldLog = *logging.Enabled
		} else {
			if encryption := conf.Get().EncryptionKeys; encryption != nil {
				if encryption.WebhookLogKey != nil {
					// Encryption enabled, default to no.
					shouldLog = false
				}
			}
		}

		if !shouldLog {
			next.ServeHTTP(w, r)
			return
		}

		// Split the body reader so we can also access it. We need to shim an
		// io.ReadCloser implementation around the TeeReader, since TeeReader
		// doesn't implement io.Closer.
		type readCloser struct {
			io.Reader
			io.Closer
		}
		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		r.Body = readCloser{tee, r.Body}

		// Put a setter into the context that can be used by
		// SetExternalServiceID to receive the external service ID from the
		// handler.
		var externalServiceID *int64
		var setter contextFunc = func(id int64) {
			externalServiceID = &id
		}
		ctx := context.WithValue(r.Context(), setterContextKey, setter)

		// Delegate to the next handler.
		next.ServeHTTP(w, r.WithContext(ctx))

		// Write the payload.
		if err := mw.store.Create(r.Context(), &types.WebhookLog{
			ExternalServiceID: externalServiceID,
			Request: types.WebhookLogRequest{
				Headers: r.Header,
				Body:    buf.Bytes(),
			},
			Error: new(string),
		}); err != nil {
			log15.Warn("error writing webhook log", "err", err)
		}
	})
}
