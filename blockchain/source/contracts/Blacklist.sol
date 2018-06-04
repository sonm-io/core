pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract Blacklist is Ownable {

    modifier OnlyMarket() {
        require(market != 0x0);
        require(msg.sender == market || master[msg.sender]);
        _;
    }

    event AddedToBlacklist(address indexed adder, address indexed addee);

    event RemovedFromBlacklist(address indexed remover, address indexed removee);

    mapping(address => mapping(address => bool)) blacklisted;

    mapping(address => bool) master;

    address public market = 0x0;

    constructor() public {
        owner = msg.sender;
    }

    function Check(address _who, address _whom) public view returns (bool) {
        return blacklisted[_who][_whom];
    }

    function Add(address _who, address _whom) external OnlyMarket returns (bool) {
        blacklisted[_who][_whom] = true;
        emit AddedToBlacklist(_who, _whom);
        return true;
    }

    function Remove(address _whom) public returns (bool) {
        require(blacklisted[msg.sender][_whom] == true);
        blacklisted[msg.sender][_whom] = false;
        emit RemovedFromBlacklist(msg.sender, _whom);
        return true;
    }

    function AddMaster(address _root) onlyOwner public returns (bool) {
        require(master[_root] == false);
        master[_root] = true;
        return true;
    }

    function RemoveMaster(address _root) onlyOwner public returns (bool) {
        require(master[_root] == true);
        master[_root] = false;
        return true;
    }

    function SetMarketAddress(address _market) onlyOwner public returns (bool) {
        market = _market;
        return true;
    }
}
