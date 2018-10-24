pragma solidity ^0.4.23;

import "./FixedPointUtil.sol";

// Common wrapper that can be used in `use FixedPoint128 for uint256` directive
library FixedPoint128 {
    uint8 constant precision = 128;

    function Round(uint256 value) public pure returns (uint256) {
        return FixedPointUtil.Round(value, precision);
    }

    function ToFixedPoint(uint256 natural) public pure returns (uint256) {
        return FixedPointUtil.FromNatural(natural, precision);
    }

    function FromNatural(uint256 natural) public pure returns (uint256) {
        return FixedPointUtil.FromNatural(natural, precision);
    }

    function FPAdd(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Add(lhs, rhs);
    }

    function FPSub(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Sub(lhs, rhs);
    }

    function FPMul(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Mul(lhs, rhs, precision);
    }

    function FPDiv(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Div(lhs, rhs, precision);
    }

    function Precision() public pure returns (uint8) {
        return precision;
    }

    function FPPow(uint256 value, uint256 power) public pure returns (uint256) {
        return FixedPointUtil.Ipow(value, power, precision);
    }
}
