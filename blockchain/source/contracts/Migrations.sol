pragma solidity ^0.4.23;
import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract Migrations is Ownable {
    uint public last_completed_migration;

    constructor() public {
        owner = msg.sender;
    }

    function setCompleted(uint completed) public onlyOwner {
        last_completed_migration = completed;
    }

    function upgrade(address new_address) public onlyOwner {
        Migrations upgraded = Migrations(new_address);
        upgraded.setCompleted(last_completed_migration);
    }
}
