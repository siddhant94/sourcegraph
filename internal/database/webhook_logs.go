package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type WebhookLogsStore struct {
	*basestore.Store
	key encryption.Key
}

func WebhookLogs(db dbutil.DB, key encryption.Key) *WebhookLogsStore {
	return &WebhookLogsStore{
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
		key:   key,
	}
}

func WebhookLogsWith(other basestore.ShareableStore, key encryption.Key) *WebhookLogsStore {
	return &WebhookLogsStore{
		Store: basestore.NewWithHandle(other.Handle()),
		key:   key,
	}
}

func (s *WebhookLogsStore) Create(ctx context.Context, log *types.WebhookLog) error {
	if mock := Mocks.WebhookLogs.Create; mock != nil {
		return mock(ctx, log)
	}

	rawRequest, err := json.Marshal(&log.Request)
	if err != nil {
		return errors.Wrap(err, "marshalling request data")
	}

	encKeyID := ""
	if s.key != nil {
		encKeyID, err = keyID(ctx, s.key)
		if err != nil {
			return errors.Wrap(err, "getting key version")
		}

		rawRequest, err = s.key.Encrypt(ctx, rawRequest)
		if err != nil {
			return errors.Wrap(err, "encrypting request data")
		}
	}

	q := sqlf.Sprintf(
		webhookLogsCreateQueryFmtstr,
		dbutil.NullInt64{N: log.ExternalServiceID},
		rawRequest,
		encKeyID,
		dbutil.NullString{S: log.Error},
		sqlf.Join(webhookLogsColumns, ", "),
	)

	row := s.QueryRow(ctx, q)
	if err := s.scanWebhookLog(ctx, log, row); err != nil {
		return errors.Wrap(err, "scanning webhook log")
	}

	return nil
}

func (s *WebhookLogsStore) GetByID(ctx context.Context, id int64) (*types.WebhookLog, error) {
	if mock := Mocks.WebhookLogs.GetByID; mock != nil {
		return mock(ctx, id)
	}

	q := sqlf.Sprintf(
		webhookLogsGetByIDQueryFmtstr,
		sqlf.Join(webhookLogsColumns, ", "),
		id,
	)

	row := s.QueryRow(ctx, q)
	log := types.WebhookLog{}
	if err := s.scanWebhookLog(ctx, &log, row); err != nil {
		return nil, errors.Wrap(err, "scanning webhook log")
	}

	return &log, nil
}

type WebhookLogsListOpts struct {
	// The maximum number of entries to return, and the cursor, if any. This
	// doesn't use LimitOffset because we're paging down a potentially changing
	// result set, so our cursor needs to be based on the ID and not the row
	// number.
	Limit  int
	Cursor int64

	// If set and non-zero, this limits the webhook logs to those matched to
	// that external service. If set and zero, this limits the webhook logs to
	// those that did not match an external service. If nil, then all webhook
	// logs will be returned.
	ExternalServiceID *int64

	// If set, only webhook logs that resulted in errors will be returned.
	OnlyErrors bool

	Since *time.Time
	Until *time.Time
}

func (opts *WebhookLogsListOpts) predicates() []*sqlf.Query {
	preds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if id := opts.ExternalServiceID; id != nil {
		if *id == 0 {
			preds = append(preds, sqlf.Sprintf("external_service_id IS NULL"))
		} else {
			preds = append(preds, sqlf.Sprintf("external_service_id = %s", *id))
		}
	}
	if opts.OnlyErrors {
		preds = append(preds, sqlf.Sprintf("error IS NOT NULL"))
	}
	if since := opts.Since; since != nil {
		preds = append(preds, sqlf.Sprintf("received_at >= %s", *since))
	}
	if until := opts.Until; until != nil {
		preds = append(preds, sqlf.Sprintf("received_at <= %s", *until))
	}

	return preds
}

func (s *WebhookLogsStore) Count(ctx context.Context, opts WebhookLogsListOpts) (int64, error) {
	if mock := Mocks.WebhookLogs.Count; mock != nil {
		return mock(ctx, opts)
	}

	q := sqlf.Sprintf(
		webhookLogsCountQueryFmtstr,
		sqlf.Join(opts.predicates(), " AND "),
	)

	row := s.QueryRow(ctx, q)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *WebhookLogsStore) List(ctx context.Context, opts WebhookLogsListOpts) ([]*types.WebhookLog, int64, error) {
	if mock := Mocks.WebhookLogs.List; mock != nil {
		return mock(ctx, opts)
	}

	preds := opts.predicates()
	if cursor := opts.Cursor; cursor != 0 {
		preds = append(preds, sqlf.Sprintf("id <= %s", cursor))
	}

	var limit *sqlf.Query
	if opts.Limit != 0 {
		limit = sqlf.Sprintf("LIMIT %s", opts.Limit+1)
	} else {
		limit = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(
		webhookLogsListQueryFmtstr,
		sqlf.Join(webhookLogsColumns, ", "),
		sqlf.Join(preds, " AND "),
		limit,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := []*types.WebhookLog{}
	for rows.Next() {
		log := types.WebhookLog{}
		if err := s.scanWebhookLog(ctx, &log, rows); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &log)
	}

	var next int64 = 0
	if opts.Limit != 0 && len(logs) == opts.Limit+1 {
		next = logs[len(logs)-1].ID
		logs = logs[:len(logs)-1]
	}

	return logs, next, nil
}

func (s *WebhookLogsStore) DeleteStale(ctx context.Context, retention time.Duration) error {
	before := time.Now().Add(-retention)

	q := sqlf.Sprintf(
		webhookLogsDeleteStaleQueryFmtstr,
		before,
	)

	return s.Exec(ctx, q)
}

var webhookLogsColumns = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("received_at"),
	sqlf.Sprintf("external_service_id"),
	sqlf.Sprintf("request"),
	sqlf.Sprintf("encryption_key_id"),
	sqlf.Sprintf("error"),
}

const webhookLogsCreateQueryFmtstr = `
-- source: internal/database/webhook_logs.go:Create
INSERT INTO
	webhook_logs (
		received_at,
		external_service_id,
		request,
		encryption_key_id,
		error
	)
	VALUES (
		NOW(),
		%s,
		%s,
		%s,
		%s
	)
	RETURNING %s
`

const webhookLogsGetByIDQueryFmtstr = `
-- source: internal/database/webhook_logs.go:GetByID
SELECT
	%s
FROM
	webhook_logs
WHERE
	id = %s
`

const webhookLogsCountQueryFmtstr = `
-- source: internal/database/webhook_logs.go:Count
SELECT
	COUNT(id)
FROM
	webhook_logs
WHERE
	%s
`

const webhookLogsListQueryFmtstr = `
-- source: internal/database/webhook_logs.go:List
SELECT
	%s
FROM
	webhook_logs
WHERE
	%s
ORDER BY
	id DESC
%s -- LIMIT
`

const webhookLogsDeleteStaleQueryFmtstr = `
-- source: internal/database/webhook_logs.go:DeleteStale
DELETE FROM
	webhook_logs
WHERE
	received_at <= %s
`

func (s *WebhookLogsStore) scanWebhookLog(ctx context.Context, log *types.WebhookLog, sc interface {
	Scan(...interface{}) error
}) error {
	var (
		encKeyID          string
		err               string
		externalServiceID int64 = -1
		rawRequest              = []byte{}
	)

	if err := sc.Scan(
		&log.ID,
		&log.ReceivedAt,
		&dbutil.NullInt64{N: &externalServiceID},
		&rawRequest,
		&encKeyID,
		&dbutil.NullString{S: &err},
	); err != nil {
		return err
	}

	if externalServiceID != -1 {
		log.ExternalServiceID = &externalServiceID
	}
	if err != "" {
		log.Error = &err
	}

	if encKeyID != "" {
		// The record includes a field indicating the encryption key ID. We
		// don't really have a way to look up a key by ID right now, so this is
		// used as a marker of whether we should expect a key or not.
		storeKeyID, err := keyID(ctx, s.key)
		if err != nil {
			return errors.Wrap(err, "retrieving store key ID")
		}

		if encKeyID != storeKeyID {
			return errors.New("key mismatch: webhook log is encrypted with a different key to the one in the store")
		}

		raw, err := s.key.Decrypt(ctx, rawRequest)
		if err != nil {
			return errors.Wrap(err, "decrypting request data")
		}

		rawRequest = []byte(raw.Secret())
	}

	if err := json.Unmarshal(rawRequest, &log.Request); err != nil {
		return errors.Wrap(err, "unmarshalling request data")
	}

	return nil
}
