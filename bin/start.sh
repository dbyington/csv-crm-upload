#!/usr/bin/env bash

DC=$(command -v docker-compose)

# Start the postgres and crm containers
${DC} up -d postgres crm
