pragma solidity ^0.4.11;


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