// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title MathLib
 * @notice Simple math library for testing weiroll value chaining
 */
contract MathLib {
    function add(uint256 a, uint256 b) external pure returns (uint256) {
        return a + b;
    }

    function multiply(uint256 a, uint256 b) external pure returns (uint256) {
        return a * b;
    }

    function subtract(uint256 a, uint256 b) external pure returns (uint256) {
        return a - b;
    }

    function divide(uint256 a, uint256 b) external pure returns (uint256) {
        require(b > 0, "Division by zero");
        return a / b;
    }

    /// @notice Extract the last element from a uint256 array
    /// @dev Useful for getting the output amount from Uniswap swaps
    function extractLastElement(uint256[] memory amounts) external pure returns (uint256) {
        require(amounts.length > 0, "Empty array");
        return amounts[amounts.length - 1];
    }
}
