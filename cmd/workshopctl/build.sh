#!/bin/sh

VERSION=$(git describe --always --dirty --long)
BUILDTIME=$(date -u '+%Y-%m-%d %H:%M:%S.%N %z %Z')
GITHASH=$(git rev-parse HEAD)

go build -ldflags " \
    -X 'github.com/joelanford/workshop/cmd/workshopctl/app.version=${VERSION}' \
    -X 'github.com/joelanford/workshop/cmd/workshopctl/app.buildTime=${BUILDTIME}' \
    -X 'github.com/joelanford/workshop/cmd/workshopctl/app.buildUser=${USER}' \
    -X 'github.com/joelanford/workshop/cmd/workshopctl/app.gitHash=${GITHASH}' \
    "