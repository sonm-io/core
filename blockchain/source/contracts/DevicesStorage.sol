pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract DevicesStorage is Ownable {

    // DATA
    struct Record {
        bytes devices;
        uint timestamp;
    }

    mapping (address => Record) devicesMap;

    bytes32 constant emptyStringHash = keccak256("");

    // EVENTS
    event DevicesHasSet(address indexed owner);
    event DevicesUpdated(address indexed owner);
    event DevicesTimestampUpdated(address indexed owner);
    event Suicide(uint block);

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

    function Touch(bytes32 _hash) public returns(bool) {
        bytes32 recordHash = keccak256(abi.encodePacked(devicesMap[msg.sender].devices));
        if (recordHash == _hash && recordHash != emptyStringHash) {
            devicesMap[msg.sender].timestamp = block.timestamp;
            emit DevicesTimestampUpdated(msg.sender);
            return true;
        }
        revert();
    }

    function Hash(address _owner) public view returns(bytes32) {
        return keccak256(abi.encodePacked(devicesMap[_owner].devices));
    }

    function GetDevices(address _owner) public view returns (bytes devices, uint timestamp) {
        Record memory record = devicesMap[_owner];
        return (record.devices, record.timestamp);
    }

    function KillDevicesStorage() public onlyOwner {
        emit Suicide(block.timestamp);
        selfdestruct(owner);
    }
}
