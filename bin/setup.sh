#!/usr/bin/env bash

set -a
source .env

export GO111MODULE=on

go build -o csvReader cmd/main.go
go build -o crmIntegrator crm/main.go

