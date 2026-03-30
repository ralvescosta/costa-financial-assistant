#!/bin/bash
# create-multiple-db.sh
# Creates multiple PostgreSQL databases from POSTGRES_MULTIPLE_DATABASES env var.
# env var: POSTGRES_MULTIPLE_DATABASES — newline or comma separated db names.
set -e

function create_db() {
  local database=$1
  echo "Creating database: $database"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE "$database";
    GRANT ALL PRIVILEGES ON DATABASE "$database" TO "$POSTGRES_USER";
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
  # Accept newlines or commas as separators
  for db in $(echo "$POSTGRES_MULTIPLE_DATABASES" | tr ',\n' ' '); do
    db=$(echo "$db" | xargs)  # trim whitespace
    [ -n "$db" ] && create_db "$db"
  done
fi
