pragma solidity ^0.4.4;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "./Declaration.sol";


//Raw prototype for Hub wallet contract.
// TODO: Punishment function - Done but not cheked
// TODO: Structure - done
// TODO: README - Done
// TODO: Registred Appendix
// TODO: Whitelist;
contract MinerWallet is Ownable {

    modifier onlyDao()     {if (msg.sender != DAO) throw; _;}

    address public DAO;

    address public Factory;

    //address public Whitelist;
    whitelist Whitelist;


    token public sharesTokenAddress;

    //uint public freezePercent;

    // FreezeQuote - it is defined amount of tokens need to be frozen on  this contract.
    uint public freezeQuote;

    uint public frozenFunds;

    uint public stakeShare;

    //lockedFunds - it is lockedFunds in percentage, which will be locked for every payday period.

    uint public lockedFunds = 0;

    //TIMELOCK
    uint64 public frozenTime;

    uint public freezePeriod;

    uint64 public genesisTime;

    //Fee's
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
    event pulledMoney(address hub, uint amount);

    ///@dev constructor
    function MinerWallet(address _minowner, address _dao, whitelist _whitelist, token sharesAddress){
        owner = _minowner;
        DAO = _dao;
        Whitelist = whitelist(_whitelist);
        Factory = msg.sender;
        genesisTime = uint64(now);

        sharesTokenAddress = token(sharesAddress);

        //1 SNM token is needed to registrate in whitelist
        freezeQuote = 1 * (1 ether / 1 wei);


        //in promilles
        daoFee = 5;

        // time of work period.
        freezePeriod = 5 days;

    }

    function Registration(uint stake) public onlyOwner returns (bool success){
        if (currentPhase != Phase.Idle) throw;
        if (sharesTokenAddress.balanceOf(this) <= freezeQuote) throw;
        stakeShare = stake;
        frozenFunds = stake + freezeQuote;
        frozenTime = uint64(now);
        //Appendix to call register function from Whitelist contract and check it.
        // TODO add to register function frozenFunds record.
        Whitelist.RegisterMin(owner, this, frozenTime, stake);
        currentPhase = Phase.Registred;
        LogPhaseSwitch(currentPhase);
        return true;
    }

    function pullMoney(address hubwallet) public onlyOwner {
        uint val = sharesTokenAddress.allowance(hubwallet, this);
        sharesTokenAddress.transferFrom(hubwallet, this, val);
        pulledMoney(hubwallet, val);
    }


    function withdraw() public onlyOwner {

        //  if(currentPhase != Phase.Created || currentPhase!=Phase.Idle) throw;
        // Miner can get funds in any time.


        if (sharesTokenAddress.balanceOf(msg.sender) < stakeShare) throw;


        uint amount = sharesTokenAddress.balanceOf(this);
        amount = amount - stakeShare;
        sharesTokenAddress.transfer(owner, amount);

    }

    function PayDay() public onlyOwner {

        if (currentPhase != Phase.Registred) throw;

        if (now < (frozenTime + freezePeriod)) throw;

        //dao got's 0.5% in such terms.
        uint DaoCollect = frozenFunds * daoFee / 1000;
        //  DaoCollect = DaoCollect + frozenFunds;
        frozenFunds = 0;
        stakeShare = 0;

        sharesTokenAddress.transfer(DAO, DaoCollect);

        //Here need to do Unregister function
        Whitelist.UnRegisterMiner(owner, this);

        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }


    function suspect() public onlyDao {
        if (currentPhase != Phase.Registred) throw;

        lockedFunds = sharesTokenAddress.balanceOf(this);
        frozenTime = uint64(now);
        currentPhase = Phase.Suspected;
        LogPhaseSwitch(currentPhase);
        freezePeriod = 120 days;
        Whitelist.UnRegisterMiner(owner, this);
    }

    function gulag() public onlyDao {
        if (currentPhase != Phase.Suspected) throw;
        if (now < (frozenTime + freezePeriod)) throw;
        uint amount = lockedFunds;
        sharesTokenAddress.transfer(DAO, amount);
        currentPhase = Phase.Punished;
        LogPhaseSwitch(currentPhase);

    }

    function rehub() public onlyDao {
        if (currentPhase != Phase.Suspected) throw;
        lockedFunds = 0;
        frozenFunds = 0;
        stakeShare = 0;
        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }

}
