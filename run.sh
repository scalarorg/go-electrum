#!/bin/bash
start() {
    go run . start --unix-socket /tmp/electrs.sock --rpc-server 127.0.0.1:60001
}

test_vault() {
    # go test -v ./electrum/vault_client_test.go
    # -count=1 prevents caching of test results
    go test -count=1 -timeout 0 ./electrum -run ^TestElectrsClient$
}

# Run all tests
test_all() {
    go test -v ./...
}

$@
