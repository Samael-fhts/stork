#!/bin/sh
set -eu

export STORK_DATABASE_HOST="localhost"
export STORK_DATABASE_PORT="5432"
export STORK_DATABASE_USER_NAME="william"
export STORK_DATABASE_PASSWORD=""
export STORK_DATABASE_NAME="storktest"
export STORK_DATABASE_MAINTENANCE_NAME="postgres"
export STORK_DATABASE_MAINTENANCE_USER_NAME="postgres"
export STORK_DATABASE_MAINTENANCE_PASSWORD=""

(\
    cd backend/server/apps/kea && \
    go test . -bench=BenchmarkLeaseFileLoad -benchmem -count 3 \
)
