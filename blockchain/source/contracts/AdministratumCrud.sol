pragma solidity ^0.4.23;

import "./Administratable.sol";

contract AdministratumCrud is Administratable {


    // storage
    mapping(address => address) masterOf;

    mapping(address => bool) flagIsMaster;

    mapping(address => mapping(address => bool)) masterRequest;

    //maps admin into its master; alternative method
    //that's like asym cryptography, but implemented by design
    mapping(address => address) admins;

    //constructor
    constructor(){
        owner = msg.sender;
        administrator = msg.sender;
    }

    function SetMaster(address _worker, address _master) public onlyOwner {
        masterOf[_worker] = _master;
    }

    function GetMaster(address _worker) public view returns (address) {
        return masterOf[_worker];
    }

    function DeleteMaster(address _worker) public onlyOwner {
        delete masterOf[_worker];
    }

    function SetAdmin(address _admin, address _master) public onlyOwner {
        admins[_admin] = _master;
    }

    function GetAdminMaster(address _admin) public view returns (address) {
        return admins[_admin];
    }

    function DeleteAdmin(address _admin) public onlyOwner {
        delete admins[_admin];
    }

    function FlagAsMaster(address _target) public onlyOwner {
        flagIsMaster[_target] = true;
    }

    function IsMaster(address _address) public view returns (bool) {
        return flagIsMaster[_address];
    }

    function UnflagAsMaster(address _target) public onlyOwner {
        delete flagIsMaster[_target];
    }
}
