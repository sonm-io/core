pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract Administratum is Ownable{
    // events


    event WorkerAnnounced(address indexed worker, address indexed master);
    event WorkerConfirmed(address indexed worker, address indexed master);
    event WorkerRemoved(address indexed worker, address indexed master);
    event adminAdded(address indexed admin, address indexed master);

    // storage
    mapping(address => address) masterOf;

    mapping(address => bool) isMaster;

    mapping(address => mapping(address => bool)) masterRequest;

    //maps admin into its master; alternative method
    //that's like asym cryptography, but implemented by design
    mapping(address => address) admins;

    mapping(address => bool) autoPayoutFlag;

    uint public payoutSupremum;

    //constructor
    constructor(){
        owner = msg.sender;
    }

    //funcs

    function RegisterWorker(address _master) public returns (bool) {
        require(GetMaster(msg.sender) == msg.sender);
        require(isMaster[msg.sender] == false);
        require(GetMaster(_master) == _master);
        masterRequest[_master][msg.sender] = true;
        emit WorkerAnnounced(msg.sender, _master);
        return true;
    }

    function ConfirmWorker(address _worker) public returns (bool) {
        require(masterRequest[msg.sender][_worker] == true ||  IsValid(_worker));
        masterOf[_worker] = msg.sender;
        isMaster[msg.sender] = true;
        delete masterRequest[msg.sender][_worker];
        emit WorkerConfirmed(_worker, msg.sender);
        return true;
    }

    function RemoveWorker(address _worker, address _master) public returns (bool) {
        require(GetMaster(_worker) == _master && (msg.sender == _worker || msg.sender == _master));
        delete masterOf[_worker];
        emit WorkerRemoved(_worker, _master);
        return true;
    }

    function RegisterAdmin(address _admin) public returns (bool){
        require(GetMaster(msg.sender) == msg.sender);
        require(msg.sender != _admin);
        admins[_admin] = msg.sender;
        return true;
    }

    function EnableAutoPayout() public returns (bool){
        require(GetMaster(msg.sender) == msg.sender);
        autoPayoutFlag[msg.sender] = true;
        return true;
    }

    function DisableAutoPayout() public returns (bool){
        require(autoPayoutFlag[msg.sender] == true);
        autoPayoutFlag[msg.sender] = false;
        return true;
    }


    //INTERNAL
    // check if transaction sended by valid admin
    function IsValid(address _worker) view internal returns(bool){
        address master = admins[msg.sender];
        return admins[msg.sender] != address(0) && masterRequest[master][_worker] == true;
    }


    //GETTERS
    function GetMaterOfAdmin(address _admin) view public returns (address){
        return admins[_admin];
    }

    function GetMaster(address _worker) view public returns (address master) {
        if (masterOf[_worker] == 0x0 || masterOf[_worker] == _worker) {
            master = _worker;
        } else {
            master = masterOf[_worker];
        }
    }

    function GetAutoPayoutFlag(address _master) view public returns (bool) {
        return autoPayoutFlag[_master];
    }

    //modifiers


}
