#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER "$POSTGRES_CSV_USER" PASSWORD '$POSTGRES_CSV_PASSWORD';
	CREATE DATABASE crm;
	GRANT ALL PRIVILEGES ON DATABASE crm TO "$POSTGRES_CSV_USER";
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_CSV_USER" --dbname "crm" <<-EOSQL
	CREATE TABLE customers (
	    -- There is an id supplied with the CSV data, so we'll use that but ensure it is supplied and unique.
	    id INTEGER NOT NULL UNIQUE,
      first_name TEXT,
      last_name TEXT,
      email TEXT NOT NULL UNIQUE,
      phone TEXT,
      -- These fields should not normally be supplied in and INSERT so they are set to the default.
      uploaded BOOLEAN NOT NULL DEFAULT false,
      created_ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      modified_ts TIMESTAMPTZ NOT NULL DEFAULT NOW());

  -- Since we'll be using the uploaded field to select rows needing to be updated create an index on it increase the
  -- speed of the select, eventhough it will be negligable with such a small data set.
  CREATE INDEX upload_idx ON customers (uploaded);

  -- This fuction and the trigger that follows will provide us with automatic updates to the modified_ts timestamp.
  CREATE FUNCTION update_modified_ts()
  RETURNS TRIGGER AS \$\$
  BEGIN
      NEW.modified_ts = NOW();
      RETURN NEW;
  END;
  \$\$ language 'plpgsql';

  CREATE TRIGGER update_modify_customers_time BEFORE UPDATE ON customers FOR EACH ROW EXECUTE PROCEDURE update_modified_ts();

EOSQL
