# PostgreSQL Scratchpad

This is a simple example that allows you to connect to a PG cluster using the official PostgreSQL client.
It enables an instant connection via `psql` by setting the [libpq environment variables](https://www.postgresql.org/docs/current/libpq-envars.html).

## Usage

1) Run the Acorn `acorn run -n pg-scratchpad .`
2) Wait for everything to finish provisioning
3) Exec into the container `acorn exec pg-scratchpad`
4) Connect to the database `psql`
5) Run queries. A simple `select 1;` will verify that you're actually connected.
