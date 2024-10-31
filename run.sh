#!/bin/bash

test_vault() {
    # go test -v ./electrum/vault_client_test.go
    # -count=1 prevents caching of test results
    go test -count=1 -timeout 0 ./electrum -run ^TestElectrsClient$
}

# Run all tests
all() {
    go test -v ./...
}

$@
