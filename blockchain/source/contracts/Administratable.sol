pragma solidity ^0.4.23;

contract Administratable {
    address public owner;
    address public administrator;


    event OwnershipRenounced(address indexed previousOwner);
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);
    event AdministratorshipTransferred(address indexed previousAdministrator, address indexed newAdministrator);

    constructor() public {
        owner = msg.sender;
        administrator = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner);
        _;
    }

    modifier ownerOrAdministrator() {
        require(msg.sender == owner || msg.sender == administrator);
        _;
    }

    modifier onlyAdministrator() {
        require(msg.sender == administrator);
        _;
    }

    function renounceOwnership() public ownerOrAdministrator {
        emit OwnershipRenounced(owner);
        owner = address(0);
    }

    function transferOwnership(address _newOwner) public ownerOrAdministrator {
        _transferOwnership(_newOwner);
    }

    function _transferOwnership(address _newOwner) internal {
        require(_newOwner != address(0));
        emit OwnershipTransferred(owner, _newOwner);
        owner = _newOwner;
    }

    function transferAdministratorship(address _newAdministrator) public onlyAdministrator {
        _transferAdministratorship(_newAdministrator);
    }

    function _transferAdministratorship(address _newAdministrator) internal {
        require(_newAdministrator != address(0));
        emit AdministratorshipTransferred(administrator, _newAdministrator);
        administrator = _newAdministrator;
    }
}
