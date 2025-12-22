# go-weiroll

A Go implementation of the [weiroll](https://github.com/weiroll/weiroll) command planner for Ethereum smart contract operation chaining.

## Overview

Weiroll is a virtual machine that batches Ethereum operations via a scripted command sequence. This library allows you to:

- Chain multiple contract calls into a single atomic transaction
- Pass return values between operations without separate transactions
- Support DELEGATECALL (libraries), CALL, STATICCALL, and CALL_WITH_VALUE

## Installation

```bash
go get github.com/branched-services/go-weiroll@latest
```

Or pin to a specific version:

```bash
go get github.com/branched-services/go-weiroll@v0.0.1
```

## Quick Start

```go
package main

import (
    "math/big"
    "github.com/branched-services/go-weiroll"
    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // Parse contract ABIs
    mathABI := weiroll.MustParseABI(mathABIJSON)
    tokenABI := weiroll.MustParseABI(tokenABIJSON)

    // Wrap contracts
    mathLib := weiroll.NewLibrary(mathLibAddr, mathABI)    // DELEGATECALL
    token := weiroll.NewContract(tokenAddr, tokenABI)      // CALL

    // Build plan
    planner := weiroll.New()

    sum := planner.Add(mathLib.MustInvoke("add", big.NewInt(1), big.NewInt(2)))
    product := planner.Add(mathLib.MustInvoke("multiply", sum, big.NewInt(10)))
    planner.Add(token.MustInvoke("transfer", recipient, product))

    // Compile
    plan, err := planner.Plan()
    if err != nil {
        panic(err)
    }

    // Execute via weiroll VM contract
    commands := plan.CommandsAsBytes32()
    state := plan.StateAsBytes()
}
```

## Features

### Contract Types

- **Library**: Contracts called via DELEGATECALL, executing in the VM's context
- **External**: Contracts called via CALL (or STATICCALL with options)

```go
// Library contract (DELEGATECALL)
lib := weiroll.NewLibrary(addr, abi)

// External contract (CALL)
contract := weiroll.NewContract(addr, abi)

// External contract with STATICCALL default
readOnly := weiroll.NewContract(addr, abi, weiroll.WithStaticCalls())
```

### Call Modifiers

```go
call := contract.MustInvoke("method", args...)

// Send ETH with call
call.WithValue(big.NewInt(1e18))

// Force STATICCALL
call.Static()

// Wrap multi-return as bytes
call.RawReturn()
```

### Value Types

```go
// Literals are created automatically from Go values
planner.Add(contract.MustInvoke("method", big.NewInt(100)))

// Or explicitly
weiroll.Uint256(big.NewInt(100))
weiroll.Address(common.Address{})
weiroll.Bytes32(common.Hash{})
weiroll.Bool(true)
weiroll.String("hello")
weiroll.Bytes([]byte{1, 2, 3})

// Return values from previous commands
sum := planner.Add(math.MustInvoke("add", 1, 2))
planner.Add(math.MustInvoke("multiply", sum, 3))  // uses sum
```

### Plan Options

```go
plan, err := planner.Plan(
    weiroll.WithSlotOptimization(true),   // Enable slot recycling (default)
    weiroll.WithMaxCommands(256),          // Max command limit
)
```

## Command Encoding

Commands are encoded as 32-byte (standard) or 64-byte (extended for >6 args) packed structures:

**Standard (≤6 args):**
```
[selector:4][flags:1][arg0-5:6][return:1][address:20]
```

**Extended (>6 args):**
```
Word 1: [selector:4][flags|0x40:1][padding:7][address:20]
Word 2: [arg slots padded to 32 bytes]
```

## State Management

The planner automatically optimizes state usage:

- **Literal Deduplication**: Identical values share the same slot
- **Slot Recycling**: Slots are reused after their last usage
- **Max 127 Slots**: Enforced limit with clear error messages

## Requirements

- Go 1.22+
- github.com/ethereum/go-ethereum

## Versioning

This project follows [Semantic Versioning](https://semver.org/). See the [releases](https://github.com/branched-services/go-weiroll/releases) for available versions.

## Testing

```bash
go test -v ./...
```

## Examples

See the [examples](./examples) directory for more detailed usage examples.

## License

MIT

## References

- [weiroll VM (Solidity)](https://github.com/weiroll/weiroll)
- [weiroll.js (JavaScript)](https://github.com/weiroll/weiroll.js)
