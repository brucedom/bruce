#!/bin/bash

rm -f /tmp/bruce
go build -ldflags "-s -w" -o /tmp/bruce cmd/main.go

/tmp/bruce $@