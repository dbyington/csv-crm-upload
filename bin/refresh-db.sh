#!/usr/bin/env bash

# Use this shell script to recreate the Postgres database.
docker exec -it csv-crm-upload_postgres_1 psql -U postgres -c "delete from customers" crm
