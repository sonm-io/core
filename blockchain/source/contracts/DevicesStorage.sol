pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract DevicesStorage is Ownable {

    // DATA
    struct Record {
        bytes devices;
        uint timestamp;
    }

    mapping (address => Record) devicesMap;

    uint public ttl = 1 days;

    // EVENTS
    event DevicesHasSet(address indexed owner);
    event DevicesUpdated(address indexed owner);
    event DevicesTimestampUpdated(address indexed owner);

    // CONTSTRUCTOR
    constructor() public {
        owner = msg.sender;
    }

    // SETTERS
    function SetDevices(bytes _deviceList) public {
        if (keccak256(devicesMap[msg.sender].devices) != keccak256("")) {
            emit DevicesUpdated(msg.sender);
        } else {
            emit DevicesHasSet(msg.sender);
        }

        devicesMap[msg.sender].devices = _deviceList;
        devicesMap[msg.sender].timestamp = block.timestamp;
    }

    function UpdateDevicesByHash(bytes32 _hash) public returns(bool) {
        bytes32 recordHash = keccak256(abi.encodePacked(devicesMap[msg.sender].devices));
        if (recordHash == _hash && recordHash != keccak256("")) {
            devicesMap[msg.sender].timestamp = block.timestamp;
            DevicesTimestampUpdated(msg.sender);
            return true;
        }
        revert();
    }

    function SetTTL(uint _new) public onlyOwner {
        ttl = _new;
    }

    // GETTERS
    function GetDevices(address _owner) public view returns (bytes devices) {
        if (devicesMap[_owner].timestamp + ttl > block.timestamp) {
            return devicesMap[_owner].devices;
        }
        return "";
    }
}
