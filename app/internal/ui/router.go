package ui

import (
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	routeLangsIndex = "index.langs"
	routeReposIndex = "index.repos"

	routeBlob        = "blob"
	routeBuild       = "build"
	routeDefLanding  = "def.landing"
	routeRepo        = "repo"
	routeRepoBuilds  = "repo.builds"
	routeRepoLanding = "repo.landing"
	routeTree        = "tree"

	routeJobs     = "jobs"
	routeTopLevel = "toplevel" // non-repo top-level routes

	routeDefRedirectToDefLanding     = "def.redirect"
	routeDefInfoRedirectToDefLanding = "def.info.redirect"
)

var router = newRouter()

func newRouter() *mux.Router {
	m := mux.NewRouter()

	m.StrictSlash(true)

	// Special top-level routes that do NOT refer to repos.
	//
	// NOTE: Keep in sync with routePatterns.tsx. See the NOTE in that
	// file for more information.
	topLevel := []string{
		// These all omit the leading "/".
		"about",
		"about/browser-ext-faqs",
		"beta",
		"contact",
		"coverage",
		"forgot",
		"join",
		"legal",
		"login",
		"pricing",
		"privacy",
		"settings",
		"reset",
		"search",
		"security",
		"styleguide",
		"terms",
		"tools",
		"tools/editor",
		"tools/browser",
		"tools",
	}
	m.Path("/sitemap").Methods("GET").Name(routeLangsIndex)
	m.Path("/sitemap/{Lang:.*}").Methods("GET").Name(routeReposIndex)
	m.Path("/{Path:(?:jobs|careers)}").Methods("GET").Name(routeJobs)
	m.Path("/{Path:(?:" + strings.Join(topLevel, "|") + ")}").Methods("GET").Name(routeTopLevel)

	// Repo
	repoPath := "/" + routevar.Repo
	repo := m.PathPrefix(repoPath + "/" + routevar.RepoPathDelim).Subrouter()
	repo.Path("/builds").Methods("GET").Name(routeRepoBuilds)
	repo.Path(`/builds/{Build:\d+}`).Methods("GET").Name(routeBuild)
	repo.Path("/info").Methods("GET").Name(routeRepoLanding)

	// RepoRev
	repoRevPath := repoPath + routevar.RepoRevSuffix
	m.Path(repoRevPath).Methods("GET").Name(routeRepo)
	repoRev := m.PathPrefix(repoRevPath + "/" + routevar.RepoPathDelim).Subrouter()
	repoRev.Path("/tree{Path:.*}").Methods("GET").Name(routeTree)
	repoRev.Path("/blob{Path:.*}").Methods("GET").Name(routeBlob)

	// Def
	repoRev.Path("/{dummy:def|refs}/" + routevar.Def).Methods("GET").Name(routeDefRedirectToDefLanding)
	repoRev.Path("/info/" + routevar.Def).Methods("GET").Name(routeDefLanding)

	return m
}
