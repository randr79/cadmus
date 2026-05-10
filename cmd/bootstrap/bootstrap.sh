#!/bin/bash

test -f ../../cadmus || (go run ./cmd/bootstrap/bootstrap.sh && go build ../..)