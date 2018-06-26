# go-solidity-sha3

> Generate Solidity SHA3 (Keccak256) hashes in Go.

[![License](http://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/miguelmota/go-solidity-sha3/master/LICENSE.md) [![Go Report Card](https://goreportcard.com/badge/github.com/miguelmota/go-solidity-sha3?)](https://goreportcard.com/report/github.com/miguelmota/go-solidity-sha3) [![GoDoc](https://godoc.org/github.com/miguelmota/go-solidity-sha3?status.svg)](https://godoc.org/github.com/miguelmota/go-solidity-sha3)

This package is the Go equivalent of `require('ethereumjs-abi').soliditySHA3` [NPM module](https://www.npmjs.com/package/ethereumjs-abi).

# Install

```bash
go get -u github.com/miguelmota/go-solidity-sha3
```

# Documentation

[Documentation on GoDoc](https://godoc.org/github.com/miguelmota/go-solidity-sha3)

# Usage

Simple example

```go
package main

import (
  "encoding/hex"
  "fmt"
  "github.com/miguelmota/go-solidity-sha3"
  "math/big"
)

func main() {
  hash := solsha3.SoliditySHA3(
    solsha3.Address("0x12459c951127e0c374ff9105dda097662a027093"),
    solsha3.Uint256(big.NewInt(100)),
    solsha3.String("foo"),
    solsha3.Bytes32("bar"),
    solsha3.Bool(true),
  )

  fmt.Println(hex.EncodeToString(hash))
}
```

Output

```bash
417a4c44724701ba79bb363151dff48909bc058a2c75a81e9cf5208ae4699369
```

# License

MIT
