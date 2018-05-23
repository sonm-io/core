pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/token/ERC20/StandardToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract SNM is StandardToken, Ownable {

    using SafeMath for uint256;

    string public name = "SONM token";

    string public symbol = "SNM";

    uint public decimals = 18;

    constructor() public {
        owner = msg.sender;
        totalSupply_ = 444 * 1e6 * 1e18;
        balances[msg.sender] = totalSupply_;
    }
}
