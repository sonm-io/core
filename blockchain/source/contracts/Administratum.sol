pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "./AdministratumCrud.sol";

contract Administratum is Ownable {

    // events
    event WorkerAnnounced(address indexed worker, address indexed master);
    event WorkerConfirmed(address indexed worker, address indexed master);
    event WorkerRemoved(address indexed worker, address indexed master);
    event AdminAdded(address indexed admin, address indexed master);

    // storage

    mapping(address => mapping(address => bool)) masterRequest;

    AdministratumCrud crud;

    // exclusivly for old api
    address market;


    //constructor
    constructor(address _administratumCrud){
        owner = msg.sender;
        crud = AdministratumCrud(_administratumCrud);
    }

    //funcs

    function RegisterWorker(address _master) public returns (bool) {
        require(GetMaster(msg.sender) == msg.sender);
        require(!crud.IsMaster(msg.sender));
        require(GetMaster(_master) == _master);
        masterRequest[_master][msg.sender] = true;
        emit WorkerAnnounced(msg.sender, _master);
        return true;
    }

    function ConfirmWorker(address _worker) public returns (bool) {
        require(masterRequest[msg.sender][_worker] == true || IsValid(_worker));
        crud.SetMaster(_worker, msg.sender);
        crud.FlagAsMaster(msg.sender);
        delete masterRequest[msg.sender][_worker];
        emit WorkerConfirmed(_worker, GetMaster(_worker));
        return true;
    }

    function RemoveWorker(address _worker, address _master) public returns (bool) {
        require(GetMaster(_worker) == _master && (msg.sender == _worker || msg.sender == _master));
        crud.DeleteMaster(_worker);
        emit WorkerRemoved(_worker, _master);
        return true;
    }

    function Migrate (address _newAdministratum) public onlyOwner {
        crud.transferOwnership(_newAdministratum);
        suicide(msg.sender);
    }

    //INTERNAL
    // check if transaction sended by valid admin
    function IsValid(address _worker) internal view returns(bool){
        address master = crud.GetAdminMaster(msg.sender);
        return master != address(0) && masterRequest[master][_worker] == true;
    }


    //GETTERS

    function GetMaster(address _worker) public view returns (address) {
        if (crud.IsMaster(_worker)) {
            return _worker;
        }

        address master = crud.GetMaster(_worker);
        if (master == address(0)) {
            return _worker;
        }

        return master;
    }

    // EXTERNAL/OLD API

    modifier OnlyMarket() {
        require(msg.sender == market);
        _;
    }

    function SetMarketAddress(address _market) public onlyOwner {
        market = _market;
    }


    function ExternalRegisterWorker(address _master, address _worker) external OnlyMarket returns (bool) {
        require(GetMaster(_worker) == _worker);
        require(!crud.IsMaster(_worker));
        require(GetMaster(_master) == _master);
        masterRequest[_master][_worker] = true;
        emit WorkerAnnounced(_worker, _master);
        return true;
    }

    function ExternalConfirmWorker(address _worker, address _master)  external OnlyMarket returns (bool) {
        require(masterRequest[_master][_worker] == true);
        crud.SetMaster(_worker, _master);
        crud.FlagAsMaster(_master);
        delete masterRequest[_master][_worker];
        emit WorkerConfirmed(_worker, _master);
        return true;
    }

    function ExternalRemoveWorker(address _worker, address _master, address _sender) external  OnlyMarket returns (bool) {
        require(GetMaster(_worker) == _master && (_sender == _worker || _sender == _master));
        crud.DeleteMaster(_worker);
        emit WorkerRemoved(_worker, _master);
        return true;
    }

    //modifiers


}
