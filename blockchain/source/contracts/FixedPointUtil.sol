pragma solidity ^0.4.23;

library FixedPointUtil {

    function FromNatural(uint256 natural, uint8 precision) internal pure returns (uint256) {
        require(natural <= ((2**256 - 1) >> uint256(precision)));
        return natural << precision;
    }

    function Round(uint256 value, uint8 precision) internal pure returns (uint256) {
        require(precision > 0);
        if(value % (uint256(2)**precision) >= (uint256(2)**(precision - 1))) {
            return (value >> precision) + 1;
        }
        return value >> precision;
    }

    function Add(uint256 lhs, uint256 rhs) internal pure returns (uint256) {
        require(2**256-1 - lhs > rhs);
        return rhs + lhs;
    }

    function Sub(uint256 lhs, uint256 rhs) internal pure returns (uint256) {
        require(lhs > rhs);
        return lhs - rhs;
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
        revert();
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
        return (lhs/rhs) << (2 * precision - lhsPrecision);
    }

    // Exponentiation by squaring algorithm
    function Ipow(uint256 value, uint256 exponent, uint8 precision) internal pure returns (uint256) {
        uint256 real_result = FromNatural(1, precision);

        while (exponent != 0) {
            if ((exponent & uint256(0x1)) == 0x1) {
                real_result = Mul(real_result, value, precision);
            }
            exponent = exponent >> 1;
            value = Mul(value, value, precision);
        }

        return real_result;
    }
}
