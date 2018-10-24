pragma solidity ^0.4.23;
import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract CreatorOwnable is Ownable {
    address creator;

    constructor() public {
        creator = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner || msg.sender == creator);
        _;
    }
}
