#!/bin/bash

set -eu

function create_database() {
  local database="$1"
  local username="$2"
  local password="$3"

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
  CREATE ROLE $username WITH LOGIN ENCRYPTED PASSWORD '$password';
  CREATE DATABASE $database WITH OWNER '$username';
  GRANT ALL PRIVILEGES ON DATABASE $database TO $username;
EOSQL
}

function drop_database() {
  local database="$1"

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
  DROP DATABASE $database;
EOSQL
}

# drop_database "root"
create_database "$MOVIE_DB_NAME" "$MOVIE_DB_ADMIN_USER" "$MOVIE_DB_ADMIN_PASSWORD"