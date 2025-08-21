#!/bin/sh
set -eu

export POSTGRES_ADDR=localhost
export POSTGRES_PASSWORD="foobar2000"
PSQL="psql postgres william"

${PSQL} -t -c "select 'drop database \"'||datname||'\";' from pg_database where datistemplate=false and datname like 'storktest%' and datname <> 'storktest';" | \
    ${PSQL}
