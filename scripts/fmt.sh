#!/usr/bin/env bash

UNFORMATTED=$(gofmt -l .)

if [ -n "$UNFORMATTED" ]; then
    echo "The following files are not properly formatted:"
    echo "$UNFORMATTED"
    echo "Please run 'go fmt ./...' before committing."
    exit 1
fi
