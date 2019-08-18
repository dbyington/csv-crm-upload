#!/usr/bin/env bash

# Use this shell script to recreate the Postgres database.

DATABASE=postgres
DATADIR='postgres/data'
DC=$(command -v docker-compose)

# The easy way, stop the container, remove the volume, start the container and it'll do the rest.
if ! ${DC} stop ${DATABASE} > /dev/null 2>&1
rm -rf ${DATADIR} > /dev/null 2>&1
${DC} up -d ${DATABASE} > /dev/null 2>&1
then
  echo "OK"
else
  echo "Restart of postgres failed"
  ${DC} logs --tail=20 ${DATABASE}
fi
