#!/bin/bash
set -e

# script timeout
(
    sleep 60
    kill $PPID
) &

# try to connect to the database
while ! psql -c "SELECT 1"; do
    sleep .5
done

echo Creating database $PGDATABASE and user $NEW_PGUSER with user $PGUSER

# create the database if it doesn't exist
DB_EXISTS=$(psql -c "SELECT FROM pg_database WHERE datname='${PGDATABASE}'")
if [[ "${DB_EXISTS}" =~ "0 rows" ]]; then
  psql -c "CREATE DATABASE \"${PGDATABASE}\""
fi

# create the user if they don't exist
USER_EXISTS=$(psql -c "SELECT FROM pg_roles WHERE rolname='${NEW_PGUSER}'")
if [[ "${USER_EXISTS}" =~ "0 rows" ]]; then
  psql -c "CREATE ROLE \"${NEW_PGUSER}\" WITH LOGIN"
fi

# set the user's password
psql -c "ALTER ROLE \"${NEW_PGUSER}\" WITH ENCRYPTED PASSWORD '${NEW_PGPASSWORD}'"

# grant the user all privileges
psql -c "GRANT ALL PRIVILEGES ON DATABASE \"${PGDATABASE}\" TO \"${NEW_PGUSER}\""
