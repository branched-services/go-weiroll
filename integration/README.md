# Weiroll Integration Tests

This directory contains integration tests that verify the go-weiroll library works correctly with an actual weiroll VM deployed on Anvil.

## Prerequisites

- [Foundry](https://getfoundry.sh) (for `forge` and `anvil`)
- Go 1.22+

## Running the Tests

The easiest way is to use the provided script:

```bash
./run_test.sh
```

This will:
1. Compile the Solidity contracts with Forge
2. Start Anvil (local Ethereum node)
3. Run the Go integration tests
4. Clean up Anvil when done

## What the Test Does

The test (`TestMathValueChaining`) demonstrates weiroll's value chaining:

```
Step 1: add(5, 3)        → returns 8      (stored in slot)
Step 2: multiply(8, 10)  → returns 80     (uses result from step 1)
Step 3: subtract(80, 20) → returns 60     (uses result from step 2)
```

All three operations execute atomically in a single transaction. The weiroll VM:
1. Decodes each command
2. Calls the target contract (MathLib)
3. Stores return values in state slots
4. Passes those slots as inputs to subsequent commands

## Contracts

- `contracts/VM.sol` - A minimal weiroll VM implementation
- `contracts/MathLib.sol` - Simple math functions for testing

## Manual Testing

If you want to run manually:

```bash
# Compile contracts
forge build

# In one terminal, start Anvil
anvil --port 8545

# In another terminal, run tests
INTEGRATION_TEST=1 go test -v -run TestMathValueChaining
```

## Understanding the Output

When you run the test, you'll see:
- Deployed contract addresses
- Encoded weiroll commands (32-byte hex strings)
- Gas used for execution
- Confirmation that value chaining worked

Example output:
```
Command[0]: 0x771602f7000102ffffffff00...  # add(5, 3)
Command[1]: 0x165c4a16000004ffffffff03...  # multiply(result, 10)
Command[2]: 0x3ef5e445000300ffffffffff...  # subtract(result, 20)
Transaction successful! Gas used: 52476
```

## Extending the Tests

To test with real protocols (like Uniswap V2), you would:
1. Fork mainnet with Anvil: `anvil --fork-url $RPC_URL`
2. Use real contract addresses
3. Fund your test account with tokens
4. Execute the swap plan

See `../examples/uniswap-v2-swap/` for an example plan (without execution).
