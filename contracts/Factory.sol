pragma solidity ^0.4.8;


import './HubWallet.sol';
import './MinerWallet.sol';


//Raw prototype of wallet factory
contract Factory {

    token sharesTokenAddress;

    address dao;

    whitelist Whitelist;

    // owner => wallet
    mapping (address => address) public hubs;

    mapping (address => address) public miners;

    event LogCreate(address wallet, address owner);

    event LogCr(address owner);

    function Factory(token TokenAddress, address _dao){
        sharesTokenAddress = TokenAddress;
        dao = _dao;
    }

    modifier onlyDao(){
        if (msg.sender != dao) throw;
        _;
    }

    function changeAdresses(address _dao, whitelist _whitelist) public onlyDao {
        dao = _dao;
        Whitelist = whitelist(_whitelist);
    }

    function createHub() public returns (address) {
        address _hubowner = msg.sender;
        address hubwallet = create(_hubowner);
        hubs[_hubowner] = hubwallet;
        LogCreate(hubwallet, _hubowner);
    }

    function createMiner() public returns (address) {
        address _minowner = msg.sender;
        address minwallet = createM(_minowner);
        miners[_minowner] = minwallet;
        LogCreate(minwallet, _minowner);
    }

    function create(address _hubowner) private returns (address) {
        return address(new HubWallet(_hubowner, dao, Whitelist, sharesTokenAddress));
        LogCr(_hubowner);
    }

    function createM(address _minowner) private returns (address) {
        return address(new MinerWallet(_minowner, dao, Whitelist, sharesTokenAddress));
        LogCr(_minowner);
    }

    function HubOf(address _owner) constant returns (address _wallet) {
        return hubs[_owner];
    }

    function MinerOf(address _owner) constant returns (address _wallet) {
        return miners[_owner];
    }

}
