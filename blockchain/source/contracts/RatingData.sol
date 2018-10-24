pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/math/SafeMath.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract RatingData is Ownable {
    using SafeMath for uint256;
    using SafeMath for uint64;

    mapping(uint8 => mapping(uint8 => mapping(address => uint64))) count;
    mapping(uint8 => mapping(uint8 => mapping(address => uint256))) sum;

    mapping(uint8 => mapping(address => uint64)) lastBlockTs;

    uint256 decayValue;

    constructor(uint256 _decayValue) public {
        decayValue = _decayValue;
    }

    // Count operations
    function Count(uint8 role, uint8 outcome, address whose) public view returns (uint64) {
        return count[role][outcome][whose];
    }

    function SetCount(uint8 role, uint8 outcome, address whose, uint64 value) public onlyOwner {
        count[role][outcome][whose] = value;
    }

    function IncCount(uint8 role, uint8 outcome, address whose) public onlyOwner {
        count[role][outcome][whose] += 1;
    }

    // Sum operations
    function Sum(uint8 role, uint8 outcome, address whose) public view returns (uint256) {
        return sum[role][outcome][whose];
    }

    function SetSum(uint8 role, uint8 outcome, address whose, uint256 value) public onlyOwner {
        sum[role][outcome][whose] = value;
    }

    function IncSum(uint8 role, uint8 outcome, address whose, uint256 value) public onlyOwner {
        sum[role][outcome][whose] = sum[role][outcome][whose].add(value);
    }

    // Last block operations
    function LastBlockTs(uint8 role, address who) public view returns (uint64) {
        return lastBlockTs[role][who];
    }

    function SetLastBlockTs(uint8 role, address who, uint64 value) public onlyOwner {
        lastBlockTs[role][who] = value;
    }

    // DecayValue operations
    function DecayValue() public view returns (uint256) {
        return decayValue;
    }

    function SetDecayValue(uint256 newDecayValue) public onlyOwner {
        decayValue = newDecayValue;
    }
}
