pragma solidity ^0.4.8;


import "./Factory.sol";


contract Whitelist {

    Factory WalletsFactory;

    mapping (address => bool) public RegistredHubs;
    mapping (address => bool) public RegistredMiners;

    event RegistredHub(address indexed _owner, address wallet, uint64 indexed time);
    event RegistredMiner(address indexed _owner, address wallet, uint64 indexed time, uint indexed stake);


    function Whitelist(Factory _factory){
        WalletsFactory = Factory(_factory);
    }

    function RegisterHub(address _owner, uint64 time) public returns (bool) {

        address wallet = WalletsFactory.HubOf(_owner);
        require(wallet == msg.sender);

        RegistredHubs[wallet] = true;
        RegistredHub(_owner, wallet, time);
        return true;
    }

    function RegisterMin(address _owner, uint64 time, uint stakeShare) public returns (bool) {
        address wallet = WalletsFactory.MinerOf(_owner);
        require(wallet == msg.sender);

        RegistredMiners[wallet] = true;
        RegistredMiner(_owner, wallet, time, stakeShare);
        return true;
    }

    function UnRegisterHub(address _owner) public returns (bool) {
        address wallet = WalletsFactory.HubOf(_owner);
        require(wallet == msg.sender);

        RegistredHubs[wallet] = false;
    }

    function UnRegisterMiner(address _owner) public returns (bool) {
        address wallet = WalletsFactory.MinerOf(_owner);
        require(wallet == msg.sender);

        RegistredMiners[wallet] = false;
    }
}
