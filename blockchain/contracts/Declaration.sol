pragma solidity ^0.4.11;


/*
 * Voting token interface
 * The token is used as a voting shares
 */
contract token {
    mapping (address => uint) public balances;
    function transferFrom(address _from, address _to, uint256 _value) returns (bool success);
    function transfer(address _to, uint _value) returns (bool success);
    function balanceOf(address _owner) constant returns (uint balance);
    function approve(address _spender, uint _value) returns (bool success);
    function allowance(address _owner, address _spender) constant returns (uint remaining);
}

/*
 * Whitelist contract interface
 * The token is used as a voting shares
 */
contract whitelist {
    function RegisterHub(address _owner, address wallet, uint64 time) public returns (bool);
    function RegisterMin(address _owner, address wallet, uint64 time, uint stakeShare) public returns (bool);
    function UnRegisterHub(address _owner, address wallet) public returns (bool);
    function UnRegisterMiner(address _owner, address wallet) public returns (bool);
}