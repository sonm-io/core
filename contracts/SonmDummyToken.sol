pragma solidity ^0.4.11;


import "zeppelin-solidity/contracts/token/MintableToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


/*
 *  Sonm Dummy Token for test
 *
 */
contract SonmDummyToken is MintableToken {

    string public name = "Sonm Dummy Token";
    string public symbol = "SDT";
    uint   public decimals = 18;
    uint   public INITIAL_SUPPLY = 1000000;

    /*
     * Constructor
     * Initied ERC20 Standart token
     * Add mint functions to readability test
     *
     */
    function SonmDummyToken(address initialAccount) {
        totalSupply = INITIAL_SUPPLY;
        balances[initialAccount] = INITIAL_SUPPLY;
    }
}
