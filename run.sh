#!/bin/bash
start() {
    TX_53705_07=0000d1c900000007a56c55847cbdb5de9d7d277a5461b58385ac6705c7e4f60af8946e1a03ca444e
    TX_53717_05=0000d1d700000004651e8b5fc6dbaae5d16a30e1658ae98699ec2834ad00c94206f1108f32efe832
    TX_53897_04=0000d28900000004df622397b9240da1c3334c3e1ad432af66a6beca9ae7cb94e74322521528619e
    TX_53903_04=0000d28f00000004b32997c3254860bd022b6a6ec90005af7977cc7a67be6bba1fcc2378093a7121
    #LAST_TX=$TX_53717_05
    #LAST_TX=$TX_53705_07
    LAST_TX=$TX_53903_04
    go run . start --unix-socket /tmp/electrs.sock --rpc-server 127.0.0.1:60001 --last-vault-tx ${LAST_TX:-''}
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
