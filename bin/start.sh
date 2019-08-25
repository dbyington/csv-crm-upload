#!/usr/bin/env bash
set -a
source .env

DC=$(command -v docker-compose)

# Start the postgres and crm containers
${DC} up -d postgres crm
