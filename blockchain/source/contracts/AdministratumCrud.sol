pragma solidity ^0.4.23;

import "./Administratable.sol";

contract AdministratumCrud is Administratable {

    // events
    event WorkerAnnounced(address indexed worker, address indexed master);
    event WorkerConfirmed(address indexed worker, address indexed master);
    event WorkerRemoved(address indexed worker, address indexed master);
    event AdminAdded(address indexed admin, address indexed master);

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

    function SetAdmin(address _admin, address _master) public onlyOwner {
        admins[_admin] = _master;
    }

    function DeleteMaster(address _worker) public onlyOwner {
        delete masterOf[_worker];
    }

    function SwitchToMaster(address _target) public onlyOwner {
        flagIsMaster[_target] = true;
    }

    function GetMaster(address _worker) public view returns (address) {
        if (masterOf[_worker] == address(0) || flagIsMaster[_worker] == true){
            return _worker;
        }
        return masterOf[_worker];
    }

    function GetAdminMaster(address _admin) public view returns (address) {
        return admins[_admin];
    }

    function isMaster(address _address) public view returns (bool) {
        return flagIsMaster[_address];
    }
}
