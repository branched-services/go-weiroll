package weiroll

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrCyclicPlanner", ErrCyclicPlanner, "weiroll: cyclic planner reference detected"},
		{"ErrInvalidSubplan", ErrInvalidSubplan, "weiroll: invalid subplan configuration"},
		{"ErrSlotExhausted", ErrSlotExhausted, "weiroll: state slot limit exceeded (max 127)"},
		{"ErrTooManyArguments", ErrTooManyArguments, "weiroll: too many arguments (max 32 for extended commands)"},
		{"ErrReturnValueNotVisible", ErrReturnValueNotVisible, "weiroll: return value not visible at this point"},
		{"ErrInvalidCallType", ErrInvalidCallType, "weiroll: invalid operation for this call type"},
		{"ErrNoReturnValue", ErrNoReturnValue, "weiroll: function has no return value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("Expected error message %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestMethodNotFoundError(t *testing.T) {
	addr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	err := &MethodNotFoundError{
		Contract: addr,
		Method:   "transfer",
	}

	expected := `weiroll: method "transfer" not found in contract 0x1234567890123456789012345678901234567890`
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestArgumentError(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		innerErr := errors.New("invalid type")
		err := &ArgumentError{
			Method: "add",
			Index:  1,
			Err:    innerErr,
		}

		expected := `weiroll: argument 1 for method "add": invalid type`
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}

		// Test Unwrap
		if err.Unwrap() != innerErr {
			t.Error("Unwrap should return the inner error")
		}
	})

	t.Run("error chain with errors.Is", func(t *testing.T) {
		err := &ArgumentError{
			Method: "add",
			Index:  0,
			Err:    ErrTooManyArguments,
		}

		if !errors.Is(err, ErrTooManyArguments) {
			t.Error("errors.Is should find ErrTooManyArguments in chain")
		}
	})
}

func TestTypeMismatchError(t *testing.T) {
	err := &TypeMismatchError{
		Expected: "uint256",
		Got:      "string",
	}

	expected := "weiroll: type mismatch: expected uint256, got string"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestPlanError(t *testing.T) {
	t.Run("with method name", func(t *testing.T) {
		innerErr := errors.New("encoding failed")
		err := &PlanError{
			CommandIndex: 5,
			Method:       "transfer",
			Err:          innerErr,
		}

		expected := "weiroll: command 5 (transfer): encoding failed"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}

		if err.Unwrap() != innerErr {
			t.Error("Unwrap should return the inner error")
		}
	})

	t.Run("without method name", func(t *testing.T) {
		innerErr := errors.New("unknown error")
		err := &PlanError{
			CommandIndex: 3,
			Method:       "",
			Err:          innerErr,
		}

		expected := "weiroll: command 3: unknown error"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("error chain with errors.Is", func(t *testing.T) {
		err := &PlanError{
			CommandIndex: 0,
			Method:       "add",
			Err:          ErrSlotExhausted,
		}

		if !errors.Is(err, ErrSlotExhausted) {
			t.Error("errors.Is should find ErrSlotExhausted in chain")
		}
	})
}

func TestEncodingError(t *testing.T) {
	t.Run("with string value", func(t *testing.T) {
		innerErr := errors.New("pack failed")
		err := &EncodingError{
			Value: "test string",
			Err:   innerErr,
		}

		expected := "weiroll: encoding error for value string: pack failed"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}

		if err.Unwrap() != innerErr {
			t.Error("Unwrap should return the inner error")
		}
	})

	t.Run("with int value", func(t *testing.T) {
		innerErr := errors.New("overflow")
		err := &EncodingError{
			Value: 12345,
			Err:   innerErr,
		}

		expected := "weiroll: encoding error for value int: overflow"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("with nil value", func(t *testing.T) {
		innerErr := errors.New("nil not allowed")
		err := &EncodingError{
			Value: nil,
			Err:   innerErr,
		}

		// nil has type "<nil>"
		if err.Error() == "" {
			t.Error("Error message should not be empty")
		}
	})

	t.Run("error chain", func(t *testing.T) {
		err := &EncodingError{
			Value: []byte{1, 2, 3},
			Err:   ErrReturnValueNotVisible,
		}

		if !errors.Is(err, ErrReturnValueNotVisible) {
			t.Error("errors.Is should find ErrReturnValueNotVisible in chain")
		}
	})
}

func TestErrorsAreDistinct(t *testing.T) {
	// Ensure all sentinel errors are distinct
	sentinelErrors := []error{
		ErrCyclicPlanner,
		ErrInvalidSubplan,
		ErrSlotExhausted,
		ErrTooManyArguments,
		ErrReturnValueNotVisible,
		ErrInvalidCallType,
		ErrNoReturnValue,
	}

	for i, err1 := range sentinelErrors {
		for j, err2 := range sentinelErrors {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("Sentinel errors %d and %d should be distinct", i, j)
			}
		}
	}
}
