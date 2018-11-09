pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "./AdministratumCrud.sol";

contract Administratum is Ownable {
    // events


    event WorkerAnnounced(address indexed worker, address indexed master);
    event WorkerConfirmed(address indexed worker, address indexed master, address indexed confirmator);
    event WorkerRemoved(address indexed worker, address indexed master);
    event AdminAdded(address indexed admin, address indexed master);

    // storage

    mapping(address => mapping(address => bool)) masterRequest;

    AdministratumCrud crud;


    //constructor
    constructor(address _administratumCrud){
        owner = msg.sender;
        crud = AdministratumCrud(_administratumCrud);
    }

    //funcs

    function RegisterWorker(address _master) public returns (bool) {
        require(crud.GetMaster(msg.sender) == msg.sender);
        require(!crud.isMaster(msg.sender));
        require(crud.GetMaster(_master) == _master);
        masterRequest[_master][msg.sender] = true;
        emit WorkerAnnounced(msg.sender, _master);
        return true;
    }

    function ConfirmWorker(address _worker) public returns (bool) {
        require(masterRequest[msg.sender][_worker] == true || IsValid(_worker));
        crud.SetMaster(_worker, msg.sender);
        crud.SwitchToMaster(msg.sender);
        delete masterRequest[msg.sender][_worker];
        emit WorkerConfirmed(_worker, crud.GetMaster(_worker), msg.sender);
        return true;
    }

    function RemoveWorker(address _worker, address _master) public returns (bool) {
        require(crud.GetMaster(_worker) == _master && (msg.sender == _worker || msg.sender == _master));
        crud.DeleteMaster(_worker);
        emit WorkerRemoved(_worker, _master);
        return true;
    }

    function RegisterAdmin(address _admin) public returns (bool){
        require(GetMaster(msg.sender) == msg.sender);
        require(msg.sender != _admin);
        crud.SetAdmin(_admin, msg.sender);
        return true;
    }


    //INTERNAL
    // check if transaction sended by valid admin
    function IsValid(address _worker) internal view returns(bool){
        address master = crud.GetAdminMaster(msg.sender);
        return master != address(0) && masterRequest[master][_worker] == true;
    }


    //GETTERS

    function GetMaster(address _worker) public view returns (address master) {
        return  crud.GetMaster(_worker);
    }

    //modifiers


}
