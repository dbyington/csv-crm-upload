#!/usr/bin/env bash

DC=$(command -v docker-compose)

# Start the postgres container
${DC} up -d postgres
