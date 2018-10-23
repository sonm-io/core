pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract AutoWithdraw is Ownable {

    address gate;

    constructor() public {
        owner = msg.sender;
    }

    struct Rate {
        address to;
        uint256 threshold;
        bool on;
    }

    mapping(address => Rate) rates;

    function Set(address _to, uint256 _threshold){
        rates[msg.sender].to = _to;
        rates[msg.sender].threshold = _threshold;
        rates[msg.sender].on = true;
    }

    function On(){
        rates[msg.sender].on = true;
    }

    function Off(){
        rates[msg.sender].on = false;
    }

    function Try(address _master){
//        require(rates[msg.sender].on);
//        token.balance > rates[msg.sender].threshold
//        gate.Payout(rates[_master].to, rates[_master].threshold);
    }


}
