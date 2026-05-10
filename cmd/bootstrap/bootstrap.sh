#!/bin/sh

if -f ../../cadmus;
    # If the binary already exists, we can skip the bootstrap process.
then
    echo "Binary already exists, skipping bootstrap."
else
    echo "Binary does not exist, running bootstrap."
    go run ./cmd/bootstrap/bootstrap.go
    go build .
fi
