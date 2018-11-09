pragma solidity ^0.4.23;


import "./SNMMasterchain.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract TestnetFaucet is Ownable {
    SNMMasterchain token;

    constructor() public {
        token = new SNMMasterchain(address(this));
        owner = msg.sender;
        token.defrost();
    }

    function getTokens() public returns (bool) {
        token.mint(msg.sender, 100*1e18);
        return true;
    }

<<<<<<< 72a9ea1ec3aa153afcfd9294eae0fe47986e4804
    function mintToken(address target, uint256 mintedAmount) public onlyOwner returns (bool) {
=======
    function mintToken(address target, uint256 mintedAmount) public onlyOwner returns (bool){
>>>>>>> some style mistakes fixed
        token.mint(target, mintedAmount);
        return true;
    }

<<<<<<< 72a9ea1ec3aa153afcfd9294eae0fe47986e4804
    function() public payable{
=======
    function() public payable {
>>>>>>> some style mistakes fixed
        getTokens();
    }

    function getTokenAddress() public view returns (address){
        return address(token);
    }
}
