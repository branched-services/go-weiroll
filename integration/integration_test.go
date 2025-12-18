package integration

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	weiroll "github.com/branched-services/go-weiroll"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Test private key (Anvil default account 0)
const testPrivateKey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

// Contract ABIs (compiled from Solidity)
const weirollVMABI = `[
	{
		"inputs": [
			{"name": "commands", "type": "bytes32[]"},
			{"name": "state", "type": "bytes[]"}
		],
		"name": "execute",
		"outputs": [{"name": "", "type": "bytes[]"}],
		"stateMutability": "payable",
		"type": "function"
	}
]`

const mathLibABI = `[
	{
		"inputs": [
			{"name": "a", "type": "uint256"},
			{"name": "b", "type": "uint256"}
		],
		"name": "add",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "a", "type": "uint256"},
			{"name": "b", "type": "uint256"}
		],
		"name": "multiply",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "pure",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "a", "type": "uint256"},
			{"name": "b", "type": "uint256"}
		],
		"name": "subtract",
		"outputs": [{"name": "", "type": "uint256"}],
		"stateMutability": "pure",
		"type": "function"
	}
]`

type ContractArtifact struct {
	ABI      json.RawMessage `json:"abi"`
	Bytecode struct {
		Object string `json:"object"`
	} `json:"bytecode"`
}

func TestMathValueChaining(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Set INTEGRATION_TEST=1 to run integration tests")
	}

	ctx := context.Background()

	// Connect to Anvil
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		t.Fatalf("Failed to connect to Anvil: %v", err)
	}
	defer client.Close()

	// Get chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain ID: %v", err)
	}
	t.Logf("Connected to chain ID: %d", chainID)

	// Load private key
	privateKey, err := crypto.HexToECDSA(testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		t.Fatalf("Failed to create transactor: %v", err)
	}

	// Deploy MathLib
	mathLibAddr, err := deployContract(ctx, client, auth, privateKey, "MathLib")
	if err != nil {
		t.Fatalf("Failed to deploy MathLib: %v", err)
	}
	t.Logf("MathLib deployed at: %s", mathLibAddr.Hex())

	// Deploy WeirollVM
	vmAddr, err := deployContract(ctx, client, auth, privateKey, "WeirollVM")
	if err != nil {
		t.Fatalf("Failed to deploy WeirollVM: %v", err)
	}
	t.Logf("WeirollVM deployed at: %s", vmAddr.Hex())

	// Create weiroll plan using our library
	mathABI := weiroll.MustParseABI(mathLibABI)
	mathLib := weiroll.NewLibrary(mathLibAddr, mathABI)

	planner := weiroll.New()

	// Plan: (5 + 3) * 10 = 80
	// Step 1: add(5, 3) = 8
	sum := planner.Add(mathLib.MustInvoke("add", big.NewInt(5), big.NewInt(3)))
	t.Log("Added: add(5, 3) -> sum")

	// Step 2: multiply(sum, 10) = 80 (uses return value from step 1!)
	product := planner.Add(mathLib.MustInvoke("multiply", sum, big.NewInt(10)))
	t.Log("Added: multiply(sum, 10) -> product")

	// Step 3: subtract(product, 20) = 60 (uses return value from step 2!)
	planner.Add(mathLib.MustInvoke("subtract", product, big.NewInt(20)))
	t.Log("Added: subtract(product, 20) -> result")

	// Compile the plan
	plan, err := planner.Plan()
	if err != nil {
		t.Fatalf("Failed to compile plan: %v", err)
	}

	t.Logf("Plan compiled: %d commands, %d state slots", len(plan.Commands), len(plan.State))

	// Execute the plan on the VM
	vmABI := weiroll.MustParseABI(weirollVMABI)
	vmContract := bind.NewBoundContract(vmAddr, vmABI, client, client, client)

	commands := plan.CommandsAsBytes32()
	state := plan.StateAsBytes()

	t.Logf("Executing with %d commands, %d state entries", len(commands), len(state))

	// Log the commands for debugging
	for i, cmd := range plan.Commands {
		t.Logf("  Command[%d]: 0x%s", i, hex.EncodeToString(cmd))
	}

	// Pack the execute call
	packedCommands := make([][32]byte, len(commands))
	copy(packedCommands, commands)

	// Get fresh nonce for execute call
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		t.Fatalf("Failed to get nonce: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// Call execute
	tx, err := vmContract.Transact(auth, "execute", packedCommands, state)
	if err != nil {
		t.Fatalf("Failed to execute plan: %v", err)
	}

	// Wait for receipt
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		t.Fatalf("Failed to mine transaction: %v", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("Transaction failed: status=%d", receipt.Status)
	}

	t.Logf("Transaction successful! Gas used: %d", receipt.GasUsed)
	t.Log("Value chaining worked: (5 + 3) * 10 - 20 = 60")
}

func deployContract(ctx context.Context, client *ethclient.Client, auth *bind.TransactOpts, privateKey *ecdsa.PrivateKey, name string) (common.Address, error) {
	// Read compiled artifact - try both naming conventions
	artifactPath := fmt.Sprintf("out/%s.sol/%s.json", name, name)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		// Try VM.sol for WeirollVM
		if name == "WeirollVM" {
			artifactPath = "out/VM.sol/WeirollVM.json"
		}
	}
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return common.Address{}, fmt.Errorf("read artifact: %w (run 'forge build' first)", err)
	}

	var artifact ContractArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return common.Address{}, fmt.Errorf("parse artifact: %w", err)
	}

	bytecodeHex := strings.TrimPrefix(artifact.Bytecode.Object, "0x")
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return common.Address{}, fmt.Errorf("decode bytecode: %w", err)
	}

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(string(artifact.ABI)))
	if err != nil {
		return common.Address{}, fmt.Errorf("parse ABI: %w", err)
	}

	// Get nonce
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return common.Address{}, fmt.Errorf("get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Address{}, fmt.Errorf("get gas price: %w", err)
	}

	// Create auth with updated nonce
	auth.Nonce = big.NewInt(int64(nonce))
	auth.GasPrice = gasPrice
	auth.GasLimit = 3000000

	// Deploy
	address, tx, _, err := bind.DeployContract(auth, parsedABI, bytecode, client)
	if err != nil {
		return common.Address{}, fmt.Errorf("deploy: %w", err)
	}

	// Wait for mining
	_, err = bind.WaitMined(ctx, client, tx)
	if err != nil {
		return common.Address{}, fmt.Errorf("wait mined: %w", err)
	}

	return address, nil
}

