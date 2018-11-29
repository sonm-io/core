pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract DevicesStorage is Ownable {

    // DATA
    struct Record {
        bytes devices;
        uint timestamp;
    }

    mapping (address => Record) devicesMap;

    uint public defaultTTL = 1 days;

    bytes32 constant emptyStringHash = keccak256("");

    // EVENTS
    event DevicesHasSet(address indexed owner);
    event DevicesUpdated(address indexed owner);
    event DevicesTimestampUpdated(address indexed owner);

    constructor() public {
        owner = msg.sender;
    }

    // SETTERS
    function SetDevices(bytes _deviceList) public {
        if (keccak256(abi.encodePacked(devicesMap[msg.sender].devices)) != emptyStringHash) {
            emit DevicesUpdated(msg.sender);
        } else {
            emit DevicesHasSet(msg.sender);
        }

        devicesMap[msg.sender].devices = _deviceList;
        devicesMap[msg.sender].timestamp = block.timestamp;
    }

    function UpdateTTL(bytes32 _hash) public returns(bool) {
        bytes32 recordHash = keccak256(abi.encodePacked(devicesMap[msg.sender].devices));
        if (recordHash == _hash && recordHash != emptyStringHash) {
            devicesMap[msg.sender].timestamp = block.timestamp;
            emit DevicesTimestampUpdated(msg.sender);
            return true;
        }
        revert();
    }

    function SetDefaultTTL(uint _new) public onlyOwner {
        defaultTTL = _new;
    }

    // GETTERS
    function TTL(address _owner) public view returns (uint) {
        if (block.timestamp > devicesMap[_owner].timestamp + defaultTTL) {
            return 0;
        }
        return devicesMap[_owner].timestamp + defaultTTL - block.timestamp ;
    }

    function Hash(address _owner) public view returns(bytes32) {
        return keccak256(abi.encodePacked(devicesMap[_owner].devices));
    }

    function GetDevices(address _owner) public view returns (bytes devices) {
        if (devicesMap[_owner].timestamp + defaultTTL > block.timestamp) {
            return devicesMap[_owner].devices;
        }
        return "";
    }
}
