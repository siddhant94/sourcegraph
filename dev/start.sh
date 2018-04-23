#!/bin/bash

if [ -n "$DELVE_FRONTEND" ]; then
	export DELVE=1
	echo 'Launching frontend with delve'
	export EXEC_FRONTEND='dlv exec --headless --listen=:2345 --log'
fi

if [ -n "$DELVE_SEARCHER" ]; then
	export DELVE=1
	echo 'Launching searcher with delve'
	export EXEC_SEARCHER='dlv exec --headless --listen=:2346 --log'
fi

if [ -n "$DELVE" ]; then
	echo 'Due to a limitation in delve, debug binaries will not start until you attach a debugger.'
	echo 'See https://github.com/derekparker/delve/issues/952'
fi

set -euf -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.." # cd to repo root dir

# Temporary failure until everyone has
# migrated. https://github.com/sourcegraph/sourcegraph/pull/10844
if [[ $PWD = *"src/sourcegraph.com/"* ]]; then
    echo "FAIL: You need to migrate the sourcegraph repo to its new import path."
    echo "We are now using github.com/sourcegraph/sourcegraph as our import path."
    echo "See the instructions at https://sourcegraph.slack.com/archives/C0EPTDE9L/p1524490333000538"
    exit 1
fi

export LIGHTSTEP_INCLUDE_SENSITIVE=true
export PGSSLMODE=disable

export GITHUB_BASE_URL=http://127.0.0.1:3180
export SRC_REPOS_DIR=$HOME/.sourcegraph/repos
export DEBUG=true
export SRC_GIT_SERVERS=127.0.0.1:3178
export SEARCHER_URL=http://127.0.0.1:3181
export REPO_UPDATER_URL=http://127.0.0.1:3182
export LSP_PROXY=127.0.0.1:4388
export REDIS_MASTER_ENDPOINT=127.0.0.1:6379
export SRC_SESSION_STORE_REDIS=127.0.0.1:6379
export SRC_INDEXER=127.0.0.1:3179
export QUERY_RUNNER_URL=http://localhost:3183
export SYMBOLS_URL=http://localhost:3184
export CTAGS_COMMAND=${CTAGS_COMMAND-cmd/symbols/universal-ctags-dev}
export CTAGS_PROCESSES=1
export SRC_SYNTECT_SERVER=http://localhost:9238
export SRC_FRONTEND_INTERNAL=localhost:3090
export SRC_PROF_HTTP=
export SRC_PROF_SERVICES=$(cat dev/src-prof-services.json)
export NPM_CONFIG_LOGLEVEL=silent

# To use webpack-dev-server for auto-reloading, use:
#   export USE_WEBPACK_DEV_SERVER=1
if [ -n "${USE_WEBPACK_DEV_SERVER-}" ]; then
	export WEBPACK_DEV_SERVER_URL=http://localhost:3088
fi

export SOURCEGRAPH_CONFIG_FILE=./dev/config.json

if ! [ -z "${ZOEKT-}" ]; then
	export ZOEKT_HOST=localhost:6070
fi

# WebApp
export NODE_ENV=development

# Make sure chokidar-cli is installed in the background
npm install &

./dev/go-install.sh

# Wait for npm install if it is still running
fg &> /dev/null || true

# Increase ulimit (not needed on Windows/WSL)
type ulimit > /dev/null && ulimit -n 10000 || true

export GOREMAN=".bin/goreman -f dev/Procfile"
exec $GOREMAN start
