pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract OracleUSD is Ownable {

    uint currentPrice = 1;

    event PriceChanged(uint price);

    constructor() public{
        owner = msg.sender;
    }

    function setCurrentPrice(uint _price) public onlyOwner{
        require(_price > 0);
        currentPrice = _price;
        emit PriceChanged(_price);
    }

<<<<<<< 72a9ea1ec3aa153afcfd9294eae0fe47986e4804
    function getCurrentPrice() public view returns (uint){
=======
    function getCurrentPrice()  public view returns (uint){
>>>>>>> some style mistakes fixed
        return currentPrice;
    }
}
