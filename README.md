# go-electrum

A Golang client library for interacting with Electrum servers, specifically designed for subscribing to and monitoring Bitcoin transactions in Scalar Vault through the Scalar Relayer component.

## Overview

This library provides a Go implementation for communicating with Electrum servers using the Electrum protocol. It's primarily used in the Scalar Relayer component to monitor and manage Bitcoin transactions for Scalar Vault.

## Features

- Connect to Electrum servers
- Subscribe to Bitcoin address notifications
- Monitor transaction status
- Handle blockchain notifications
- Support for Scalar Vault transaction monitoring

## Installation

```
go get github.com/scalar-vault/go-electrum
```

## Usage

Please refer to the [test cases](./electrum/vault_client_test.go) for more examples.

## Features

- TCP connection management
- SSL/TLS support
- JSON-RPC protocol implementation
- Address subscription
- Transaction monitoring
- Error handling and reconnection logic

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License

## Related Projects

- [Scalar Bitcoin Vault](https://github.com/scalarorg/bitcoin-vault)
- [Scalar Relayer](https://github.com/scalarorg/relayer)
