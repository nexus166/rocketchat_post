#!/usr/bin/env bash

set -xeo pipefail

semver="$(git describe --tags)"
commit="$(git rev-parse --short HEAD)"
timestamp="$(date -u '+%FT%TZ')"
webhook="${SLACK_WEBHOOK}"

CGO_ENABLED=0 go build -o hermes -v -ldflags="-s -w -X main.defaultWebHook=${webhook} -X main.semver=${semver} -X main.commit=${commit} -X main.built=${timestamp}"
