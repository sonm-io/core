pragma solidity ^0.4.23;


import "./SNMMasterchain.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract TestnetFaucet is Ownable {
    SNMMasterchain token;

    constructor(address _token) public{
        token = SNMMasterchain(_token);
        owner = msg.sender;
    }

    function getTokens() public returns (bool){
        token.mint(msg.sender, 100*1e18);
    }

    function mintToken(address target, uint256 mintedAmount) onlyOwner public returns (bool){
        token.mint(target, mintedAmount);
    }

    function() payable public{
        getTokens();
    }
}
