#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Weiroll Integration Test ==="
echo ""

# Check for Foundry
if ! command -v forge &> /dev/null; then
    echo "Error: Foundry (forge) not found. Install from https://getfoundry.sh"
    exit 1
fi

if ! command -v anvil &> /dev/null; then
    echo "Error: Anvil not found. Install from https://getfoundry.sh"
    exit 1
fi

# Compile contracts
echo "1. Compiling contracts with Forge..."
forge build --silent
echo "   ✓ Contracts compiled"

# Start Anvil in background
echo "2. Starting Anvil..."
anvil --port 8545 &> /dev/null &
ANVIL_PID=$!

# Cleanup on exit
cleanup() {
    echo ""
    echo "Cleaning up..."
    kill $ANVIL_PID 2>/dev/null || true
}
trap cleanup EXIT

# Wait for Anvil to be ready
sleep 2
echo "   ✓ Anvil running (PID: $ANVIL_PID)"

# Run the Go test
echo "3. Running integration test..."
echo ""
cd "$SCRIPT_DIR"
INTEGRATION_TEST=1 go test -v -run TestMathValueChaining

echo ""
echo "=== Test Complete ==="
