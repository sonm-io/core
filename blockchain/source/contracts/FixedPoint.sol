pragma solidity ^0.4.23;

import "./FixedPointUtil.sol";

contract FixedPoint {

    uint8 precision;
    constructor(uint8 _precision) public {
        require(_precision > 0);
        precision = _precision;
    }

    function Round(uint256 value) public view returns (uint256) {
        return FixedPointUtil.Round(value, precision);
    }

    function FromNatural(uint256 natural) public view returns (uint256) {
        return FixedPointUtil.FromNatural(natural, precision);
    }

    function Add(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Add(lhs, rhs);
    }

    function Sub(uint256 lhs, uint256 rhs) public pure returns (uint256) {
        return FixedPointUtil.Sub(lhs, rhs);
    }

    function Mul(uint256 lhs, uint256 rhs) public view returns (uint256) {
        return FixedPointUtil.Mul(lhs, rhs, precision);
    }

    function Div(uint256 lhs, uint256 rhs) public view returns (uint256) {
        return FixedPointUtil.Div(lhs, rhs, precision);
    }

    function Precision() public view returns (uint8) {
        return precision;
    }

    function Ipow(uint256 value, uint256 power) public view returns (uint256) {
        return FixedPointUtil.Ipow(value, power, precision);
    }
}
