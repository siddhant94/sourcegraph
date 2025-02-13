package database

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// A Namespace is a username or an organization name. No user may have a username that is equal to
// an organization name, and vice versa. This property means that a username or organization name
// serves as a namespace for other objects that are owned by the user or organization, such as
// batch changes and extensions.
type Namespace struct {
	// Name is the canonical-case name of the namespace (which is unique among all namespace
	// types). For a user, this is the username. For an organization, this is the organization name.
	Name string

	User, Organization int32 // exactly 1 is non-zero
}

var (
	ErrNamespaceMultipleIDs = errors.New("multiple namespace IDs provided")
	ErrNamespaceNoID        = errors.New("no namespace ID provided")
	ErrNamespaceNotFound    = errors.New("namespace not found")
)

type NamespaceStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) NamespaceStore
	Transact(ctx context.Context) (NamespaceStore, error)
	GetByID(ctx context.Context, orgID, userID int32) (*Namespace, error)
	GetByName(ctx context.Context, name string) (*Namespace, error)
}

type namespaceStore struct {
	*basestore.Store
}

// Namespaces instantiates and returns a new NamespaceStore with prepared statements.
func Namespaces(db dbutil.DB) NamespaceStore {
	return &namespaceStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewNamespaceStoreWithDB instantiates and returns a new NamespaceStore using the other store handle.
func NamespacesWith(other basestore.ShareableStore) NamespaceStore {
	return &namespaceStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *namespaceStore) With(other basestore.ShareableStore) NamespaceStore {
	return &namespaceStore{Store: s.Store.With(other)}
}

func (s *namespaceStore) Transact(ctx context.Context) (NamespaceStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &namespaceStore{Store: txBase}, err
}

// GetByID looks up the namespace by an ID.
//
// One of orgID and userID must be 0: whichever ID is non-zero will be used to
// look up the namespace. If both are given, ErrNamespaceMultipleIDs is
// returned; if neither are given, ErrNamespaceNoID is returned.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func (s *namespaceStore) GetByID(
	ctx context.Context,
	orgID, userID int32,
) (*Namespace, error) {
	if Mocks.Namespaces.GetByID != nil {
		return Mocks.Namespaces.GetByID(ctx, orgID, userID)
	}

	preds := []*sqlf.Query{}
	if orgID != 0 && userID != 0 {
		return nil, ErrNamespaceMultipleIDs
	} else if orgID != 0 {
		preds = append(preds, sqlf.Sprintf("org_id = %s", orgID))
	} else if userID != 0 {
		preds = append(preds, sqlf.Sprintf("user_id = %s", userID))
	} else {
		return nil, ErrNamespaceNoID
	}

	var n Namespace
	if err := s.getNamespace(ctx, &n, preds); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetByName looks up the namespace by a name. The name is matched
// case-insensitively against all namespaces, which is the set of usernames and
// organization names.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func (s *namespaceStore) GetByName(
	ctx context.Context,
	name string,
) (*Namespace, error) {
	if Mocks.Namespaces.GetByName != nil {
		return Mocks.Namespaces.GetByName(ctx, name)
	}

	var n Namespace
	if err := s.getNamespace(ctx, &n, []*sqlf.Query{
		sqlf.Sprintf("name = %s", name),
	}); err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *namespaceStore) getNamespace(ctx context.Context, n *Namespace, preds []*sqlf.Query) error {
	q := getNamespaceQuery(preds)
	err := s.QueryRow(
		ctx,
		q,
	).Scan(&n.Name, &n.User, &n.Organization)

	if err == sql.ErrNoRows {
		return ErrNamespaceNotFound
	}
	return err
}

func getNamespaceQuery(preds []*sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(namespaceQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

var namespaceQueryFmtstr = `
-- source: internal/database/namespaces.go:getNamespace
SELECT
	name,
	COALESCE(user_id, 0) AS user_id,
	COALESCE(org_id, 0) AS org_id
FROM
	names
WHERE %s
`
