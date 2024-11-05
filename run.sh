#!/bin/bash
start() {
    #
    go run . start --unix-socket /tmp/electrs.sock --rpc-server 192.168.1.254:60001
    # go run . start --unix-socket /tmp/electrs.sock --rpc-server 18.140.72.123:60001
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
