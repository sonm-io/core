pragma solidity ^0.4.11;

import "./SonmDummyToken.sol";
import "./Whitelist.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract HubWallet is Ownable{

    Whitelist whitelist;

    function HubWallet(address _owner, Whitelist _whitelist){
        owner = owner;
        whitelist = _whitelist;
    }

}