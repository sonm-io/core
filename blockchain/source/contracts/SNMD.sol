pragma solidity ^0.4.21;


import "zeppelin-solidity/contracts/token/ERC20/StandardToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract SNMD is StandardToken, Ownable {

    using SafeMath for uint256;

    string public name = "SONM token";

    string public symbol = "SNM";

    uint public decimals = 18;

    mapping (address => bool) markets;

    uint totalSupply_ = 444 * 1e6 * 1e18;

    function SNMD() public{
        owner = msg.sender;

        balances[msg.sender] = totalSupply_;
    }

    function AddMarket(address _newMarket) onlyOwner public {
        markets[_newMarket] = true;
    }

    //overrided for market
    function transferFrom(address _from, address _to, uint256 _value) public returns (bool) {
        if(markets[_to] == true){
            require(market != address(0));
            require(markets[msg.sender] == true);
            require(_value <= balances[_from]);

            balances[_from] = balances[_from].sub(_value);
            balances[_to] = balances[_to].add(_value);
            Transfer(_from, _to, _value);

            return true;
        } else {
            require(_to != address(0));
            require(_value <= balances[_from]);
            require(_value <= allowed[_from][msg.sender]);

            balances[_from] = balances[_from].sub(_value);
            balances[_to] = balances[_to].add(_value);
            allowed[_from][msg.sender] = allowed[_from][msg.sender].sub(_value);
            Transfer(_from, _to, _value);

            return true;
        }
    }
}
