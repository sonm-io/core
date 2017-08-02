pragma solidity ^0.4.11;


import "./SonmDummyToken.sol";


contract Factory {

    SonmDummyToken public tokenAddress;

    function Factory(SonmDummyToken TokenAddress){
        tokenAddress = TokenAddress;
    }


}