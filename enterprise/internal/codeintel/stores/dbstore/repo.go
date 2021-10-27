package dbstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// RepoIDsByGlobPattern returns a slice of IDs from the repo table that matches the pattern string.
func (s *Store) RepoIDsByGlobPattern(ctx context.Context, pattern string) (_ []int, err error) {
	ctx, endObservation := s.operations.repoIDsByGlobPattern.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", pattern),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, s.Store.Handle().DB())
	if err != nil {
		return nil, err
	}

	return basestore.ScanInts(s.Store.Query(ctx, sqlf.Sprintf(repoIDsByGlobPatternQuery, strings.ReplaceAll(pattern, "*", "%"), authzConds)))
}

const repoIDsByGlobPatternQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:RepoIDsByGlobPattern
SELECT id
FROM repo
WHERE
	name ILIKE %s AND
	deleted_at IS NULL AND
	blocked IS NULL AND (%s)
`

// UpdateReposMatchingPatterns updates lsif_configuration_policies_repository_pattern_lookup table
// from the patterns set in the repository_pattern column from repo tables.
func (s *Store) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int) (err error) {
	ctx, endObservation := s.operations.updateReposMatchingPatterns.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", strings.Join(patterns, ",")),
	}})
	defer endObservation(1, observation.Args{})

	conds := make([]*sqlf.Query, 0, len(patterns))
	if len(patterns) == 0 {
		// When patterns is zero, we set the WHERE clause to FALSE
		// to make sure `repos` is empty so we can just trigger the `deleted` CTE.
		conds = append(conds, sqlf.Sprintf("FALSE"))
	} else {
		for _, pattern := range patterns {
			conds = append(conds, sqlf.Sprintf("name ILIKE %s", strings.ReplaceAll(pattern, "*", "%")))
		}
	}

	return s.Store.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatternsQuery, sqlf.Join(conds, "OR"), policyID, policyID))
}

const updateReposMatchingPatternsQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:UpdateReposMatchingPatterns
WITH
repos AS (
	SELECT id
	FROM repo
	WHERE
		deleted_at IS NULL AND
		blocked IS NULL AND
		(%s)
),
deleted AS (
	DELETE FROM lsif_configuration_policies_repository_pattern_lookup
	WHERE
		policy_id = %s AND
		-- Do not delete repos we're inserting
		repo_id NOT IN (SELECT id FROM repos)
)
INSERT INTO lsif_configuration_policies_repository_pattern_lookup(policy_id, repo_id)
SELECT %s, repos.id
FROM repos
`
