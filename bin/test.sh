#!/usr/bin/env bash
set -a
source .env
set +a
# shellcheck disable=SC2068
ginkgo -cover -race -trace $@ ./...
