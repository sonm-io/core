pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract AddressHashMap is Ownable {
    mapping(bytes32 => address) map;
    constructor(){
        owner = msg.sender;
    }

    function write(bytes32 _key, address _value) public onlyOwner {
        map[_key] = _value;
    }

    function read(bytes32 _key) public view returns (address){
        return map[_key];
    }
}
