pragma solidity ^0.4.4;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/token/ERC20.sol";
import "./Whitelist.sol";


//Raw prototype for Hub wallet contract.

contract MinerWallet is Ownable {

    modifier onlyDao()     { require(msg.sender == DAO); _; }

    address public DAO;

    address public Factory;

    Whitelist whitelist;

    ERC20 public sharesTokenAddress;

    // defined amount of tokens need to be frozen on  this contract.
    uint public freezeQuote;

    uint public frozenFunds;

    uint public stakeShare;

    // lockedFunds in percentage, which will be locked for every payday period.

    uint public lockedFunds = 0;

    //TIMELOCK
    uint64 public frozenTime;

    uint public freezePeriod;

    uint64 public genesisTime;

    //Fees
    uint daoFee;


    enum Phase {
        Created,
        Registred,
        Idle,
        Suspected,
        Punished
    }

    Phase public currentPhase = Phase.Idle;


    event LogPhaseSwitch(Phase newPhase);
    event PulledMoney(address hub, uint amount);

    ///@dev constructor
    function MinerWallet(address _minowner, address _dao, Whitelist _whitelist, ERC20 _sharesAddress){
        owner = _minowner;
        DAO = _dao;
        whitelist = _whitelist;
        Factory = msg.sender;
        genesisTime = uint64(now);
        sharesTokenAddress = _sharesAddress;

        //1 SNM token is needed to registrate in whitelist
        freezeQuote = 1 * (1 ether / 1 wei);

        daoFee = 5; // in promilles
        freezePeriod = 5 days; // time of work period.
    }

    function Registration(uint stake) public onlyOwner returns (bool success){
        require(currentPhase == Phase.Idle);
        require(sharesTokenAddress.balanceOf(this) > freezeQuote);

        stakeShare = stake;
        frozenFunds = stake + freezeQuote;
        frozenTime = uint64(now);
        //Appendix to call register function from Whitelist contract and check it.
        // TODO add to register function frozenFunds record.
        whitelist.RegisterMin(this, frozenTime, stake);
        currentPhase = Phase.Registred;
        LogPhaseSwitch(currentPhase);
        return true;
    }

    function pullMoney(address hubwallet) public onlyOwner {
        uint val = sharesTokenAddress.allowance(hubwallet, this);
        sharesTokenAddress.transferFrom(hubwallet, this, val);
        PulledMoney(hubwallet, val);
    }


    function withdraw() public onlyOwner {
        require(sharesTokenAddress.balanceOf(msg.sender) >= stakeShare);

        uint amount = sharesTokenAddress.balanceOf(this);
        amount = amount - stakeShare;
        sharesTokenAddress.transfer(owner, amount);
    }

    function PayDay() public onlyOwner {
        require(currentPhase == Phase.Registred);
        require(now >= frozenTime + freezePeriod);

        //dao got's 0.5% in such terms.
        uint DaoCollect = frozenFunds * daoFee / 1000;
        //  DaoCollect = DaoCollect + frozenFunds;
        frozenFunds = 0;
        stakeShare = 0;

        sharesTokenAddress.transfer(DAO, DaoCollect);

        //Here need to do Unregister function
        whitelist.UnRegisterMiner(this);

        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }


    function suspect() public onlyDao {
        require(currentPhase == Phase.Registred);

        lockedFunds = sharesTokenAddress.balanceOf(this);
        frozenTime = uint64(now);
        currentPhase = Phase.Suspected;
        LogPhaseSwitch(currentPhase);
        freezePeriod = 120 days;
        whitelist.UnRegisterMiner(this);
    }

    function gulag() public onlyDao {
        require(currentPhase == Phase.Suspected);
        require(now >= frozenTime + freezePeriod);

        uint amount = lockedFunds;
        sharesTokenAddress.transfer(DAO, amount);
        currentPhase = Phase.Punished;
        LogPhaseSwitch(currentPhase);
    }

    function rehub() public onlyDao {
        require(currentPhase == Phase.Suspected);

        lockedFunds = 0;
        frozenFunds = 0;
        stakeShare = 0;
        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }
}
