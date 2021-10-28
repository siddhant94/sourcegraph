package database

import (
	"context"
	"strconv"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A Cursor for efficient index based pagination through large result sets.
type Cursor struct {
	// Column contains the relevant column for cursor-based pagination (e.g. "name")
	Column string
	// Value contains the relevant value for cursor-based pagination (e.g. "Zaphod").
	Value string
	// Direction contains the comparison for cursor-based pagination, all possible values are: next, prev.
	Direction string
}

func NewRepoNamesCursor(column, value, direction string) (*Cursor, error) {
	if column == "" {
		column = "id"
	}

	// TODO(tsenart): Add support for stars cursor pagination, which will require
	// adding that field to types.RepoName.
	switch column {
	case "id", "name":
	default:
		return nil, errors.Newf("invalid cursor column %q", column)
	}

	if direction == "" {
		direction = "next"
	}

	switch direction {
	case "next", "prev":
	default:
		return nil, errors.Newf("invalid cursor direction %q", direction)
	}

	return &Cursor{Column: column, Value: value, Direction: direction}, nil
}

func NewRepoNamesIterator(s RepoStore, opts *ReposListOptions) (*RepoNamesIterator, error) {
	if s == nil {
		return nil, errors.New("NewStoreRepoNamesIterator: invalid RepoStore(nil)")
	}

	if opts == nil {
		opts = &ReposListOptions{NoPrivate: true}
	}

	if opts.Cursor == nil {
		return nil, errors.New("NewStoreRepoNamesIterator: no Cursor in ReposListOptions")
	}

	return &RepoNamesIterator{s: s, opts: *opts}, nil
}

type RepoNamesIterator struct {
	mu   sync.Mutex
	s    RepoStore
	opts ReposListOptions
}

func (it *RepoNamesIterator) Next(ctx context.Context, n int) ([]*types.RepoName, error) {
	if n <= 0 {
		return nil, nil
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	if it.opts.LimitOffset == nil {
		it.opts.LimitOffset = &LimitOffset{}
	}

	if it.opts.Limit != n {
		it.opts.Limit = n
	}

	repos := make([]*types.RepoName, 0, n)
	err := it.s.StreamRepoNames(ctx, it.opts, func(r *types.RepoName) {
		repos = append(repos, r)
	})
	if err != nil {
		return nil, err
	}

	if len(repos) > 0 {
		last := repos[len(repos)-1]
		switch it.opts.Cursor.Column {
		case "id":
			it.opts.Cursor.Value = strconv.FormatInt(int64(last.ID), 10)
		case "name":
			it.opts.Cursor.Value = string(last.Name)
		}
	}

	return repos, nil
}

type CachedRepoNamesIterator struct {
	mu    sync.Mutex
	it    RepoNamesIterator
	repos []*types.RepoName
}

func (it *CachedRepoNamesIterator) Next(ctx context.Context, n int) ([]*types.RepoName, error) {
	it.mu.Lock()
	defer it.mu.Unlock()
	repos, err := it.it.Next(ctx, n)
	it.repos = append(it.repos, repos...)
	return repos, err
}

func (it *CachedRepoNamesIterator)
