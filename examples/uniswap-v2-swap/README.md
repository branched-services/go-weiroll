# Uniswap V2 Multi-Step Swap Example

## What This Example Does

This example demonstrates **value chaining** in weiroll - the ability to pass return values from one contract call to the next, all in a single atomic transaction.

### The Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    SINGLE ATOMIC TRANSACTION                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Step 1: WETH.approve(router, MAX_UINT256)                      │
│          └─> Allows router to spend our WETH                    │
│                                                                  │
│  Step 2: router.swapExactTokensForTokens(                       │
│            amountIn: 1 WETH,                                    │
│            path: [WETH, USDC],                                  │
│            ...                                                   │
│          )                                                       │
│          └─> Returns: [1 WETH, X USDC]  ─────────┐              │
│                                                   │              │
│  Step 3: helper.extractLastElement(amounts) <────┘              │
│          └─> Returns: X USDC (the output amount)  ───┐          │
│                                                       │          │
│  Step 4: USDC.approve(router, MAX_UINT256)           │          │
│          └─> Allows router to spend our USDC         │          │
│                                                       │          │
│  Step 5: router.swapExactTokensForTokens(            │          │
│            amountIn: X USDC  <────────────────────────┘          │
│            path: [USDC, DAI],                                    │
│            ...                                                   │
│          )                                                       │
│          └─> Returns: [X USDC, Y DAI]                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Why This Matters

Without weiroll, you'd need to either:
1. **Multiple transactions**: Swap WETH→USDC, wait for confirmation, then swap USDC→DAI
2. **Hardcode amounts**: Know exactly how much USDC you'll get before submitting

With weiroll:
- Everything happens atomically (all or nothing)
- The output of one swap feeds directly into the next
- No front-running risk between steps

## Testing with Anvil

See the `integration_test.go` file for a real test against forked mainnet.
