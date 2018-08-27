pragma solidity ^0.4.23;

library FixedPointUtil {

    function Round(uint256 value, uint8 precision) internal pure returns (uint256) {
        require(precision > 0);
        if(value % (uint256(2)**precision) >= (uint256(2)**(precision - 1))) {
            return (value >> precision) + 1;
        }
        return value >> precision;
    }

    function Mul(uint256 lhs, uint256 rhs, uint8 precision) internal pure returns (uint256) {
        require(precision > 0);
        if (lhs == 0 || rhs == 0) {
            return 0;
        }

        uint8 p = precision;
        for(uint8 i = 0; i <= precision; i++) {
            uint256 c = lhs * rhs;
            if (c / lhs == rhs) {
                return c >> p;
            }
            p -= 1;
            if (lhs > rhs) {
                lhs = lhs >> 1;
            } else {
                rhs = rhs >> 1;
            }
        }
        return rhs;
//        revert();
    }

    function Div(uint256 lhs, uint256 rhs, uint8 precision) internal pure returns (uint256) {
        require(precision > 0);
        require(rhs > 0);
        uint8 lhsPrecision = precision;
        for(uint8 i = 0; i < precision; i++) {
            if (lhs >= 1 << 255) {
                break;
            }
            lhs = lhs * 2;
            lhsPrecision += 1;
        }
        return (lhs/rhs) >> (lhsPrecision - precision);
    }
}

contract FixedPoint {

    uint8 precision;
    constructor(uint8 _precision) public {
        require(_precision > 0);
        precision = _precision;
    }

    function Round(uint256 value) public view returns (uint256) {
        return FixedPointUtil.Round(value, precision);
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
}
