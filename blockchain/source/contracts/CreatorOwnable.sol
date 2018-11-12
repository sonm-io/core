pragma solidity ^0.4.23;
import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract CreatorOwnable is Ownable {
    address creator;

    constructor() public {
        creator = msg.sender;
    }

    function transferOwnership(address _newOwner) public ownerOrCreator {
        _transferOwnership(_newOwner);
    }

    modifier ownerOrCreator() {
        require(msg.sender == owner || msg.sender == creator);
        _;
    }
}
