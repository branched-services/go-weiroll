// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title Weiroll Virtual Machine
 * @notice Executes a sequence of commands that can chain return values
 * @dev Based on https://github.com/weiroll/weiroll
 */
contract WeirollVM {
    error ExecutionFailed(uint256 commandIndex, address target, bytes message);

    uint8 constant FLAG_CT_MASK = 0x03;      // Call type mask
    uint8 constant FLAG_CT_DELEGATECALL = 0x00;
    uint8 constant FLAG_CT_CALL = 0x01;
    uint8 constant FLAG_CT_STATICCALL = 0x02;
    uint8 constant FLAG_CT_VALUECALL = 0x03;
    uint8 constant FLAG_TUPLE_RETURN = 0x04;
    uint8 constant FLAG_EXTENDED_COMMAND = 0x40;

    uint8 constant SHORT_COMMAND_FILL = 0x00; // unused arg slots
    uint8 constant USE_STATE = 0xfe;          // magic value for state array ref
    uint8 constant END_OF_ARGS = 0xff;        // no more args / no return

    /**
     * @notice Execute a sequence of weiroll commands
     * @param commands Array of 32-byte encoded commands
     * @param state Initial state array (return values are written here)
     * @return Final state array after execution
     */
    function execute(bytes32[] calldata commands, bytes[] memory state)
        external
        payable
        returns (bytes[] memory)
    {
        for (uint256 i = 0; i < commands.length; i++) {
            bytes32 command = commands[i];

            // Parse command
            bytes4 selector = bytes4(command);
            uint8 flags = uint8(command[4]);
            // Address is in the last 20 bytes of the command
            address target = address(uint160(uint256(command)));

            // Handle extended commands (>6 args)
            bytes memory args;
            uint8 returnSlot;

            if (flags & FLAG_EXTENDED_COMMAND != 0) {
                // Extended command: next word contains arg slots
                i++;
                bytes32 extArgs = commands[i];
                (args, returnSlot) = _buildArgsExtended(selector, extArgs, state);
            } else {
                // Standard command: args in bytes 5-10, return in byte 11
                (args, returnSlot) = _buildArgs(selector, command, state);
            }

            // Execute call
            bool success;
            bytes memory result;
            uint8 callType = flags & FLAG_CT_MASK;

            if (callType == FLAG_CT_DELEGATECALL) {
                (success, result) = target.delegatecall(args);
            } else if (callType == FLAG_CT_CALL) {
                (success, result) = target.call(args);
            } else if (callType == FLAG_CT_STATICCALL) {
                (success, result) = target.staticcall(args);
            } else if (callType == FLAG_CT_VALUECALL) {
                // Value is in last arg slot
                uint256 value = abi.decode(state[uint8(command[10])], (uint256));
                (success, result) = target.call{value: value}(args);
            }

            if (!success) {
                revert ExecutionFailed(i, target, result);
            }

            // Store return value if needed
            if (returnSlot != END_OF_ARGS) {
                uint8 slot = returnSlot & 0x7f;
                if (flags & FLAG_TUPLE_RETURN != 0) {
                    // Store raw bytes
                    state[slot] = result;
                } else {
                    // Store decoded value
                    state[slot] = result;
                }
            }
        }

        return state;
    }

    function _buildArgs(bytes4 selector, bytes32 command, bytes[] memory state)
        internal
        pure
        returns (bytes memory, uint8 returnSlot)
    {
        // Args are in bytes 5-10, return slot in byte 11
        returnSlot = uint8(command[11]);

        bytes memory args = abi.encodePacked(selector);

        for (uint256 j = 5; j <= 10; j++) {
            uint8 slot = uint8(command[j]);
            if (slot == END_OF_ARGS) break;
            if (slot == USE_STATE) {
                // Encode entire state array
                args = abi.encodePacked(args, abi.encode(state));
            } else {
                bool isDynamic = (slot & 0x80) != 0;
                uint8 idx = slot & 0x7f;
                if (isDynamic) {
                    args = abi.encodePacked(args, state[idx]);
                } else {
                    args = abi.encodePacked(args, state[idx]);
                }
            }
        }

        return (args, returnSlot);
    }

    function _buildArgsExtended(bytes4 selector, bytes32 extArgs, bytes[] memory state)
        internal
        pure
        returns (bytes memory, uint8 returnSlot)
    {
        bytes memory args = abi.encodePacked(selector);

        uint256 j = 0;
        while (j < 32) {
            uint8 slot = uint8(extArgs[j]);
            if (slot == END_OF_ARGS) {
                // Return slot is next byte
                if (j + 1 < 32) {
                    returnSlot = uint8(extArgs[j + 1]);
                } else {
                    returnSlot = END_OF_ARGS;
                }
                break;
            }

            bool isDynamic = (slot & 0x80) != 0;
            uint8 idx = slot & 0x7f;
            if (isDynamic) {
                args = abi.encodePacked(args, state[idx]);
            } else {
                args = abi.encodePacked(args, state[idx]);
            }
            j++;
        }

        return (args, returnSlot);
    }

    // Allow receiving ETH
    receive() external payable {}
}
