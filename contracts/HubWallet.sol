pragma solidity ^0.4.4;


//Raw prototype for Hub wallet contract.

// TODO: Punishment function - Done but not cheked
// TODO: Structure - done
// TODO: README - Done
// TODO: Registred Appendix
// TODO: Whitelist;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "./Declaration.sol";


contract HubWallet is Ownable {

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

    //lockedFunds - it is lockedFunds in percentage, which will be locked for every payday period.
    uint public lockPercent;

    uint public lockedFunds = 0;

    //TIMELOCK
    uint64 public frozenTime;

    uint public freezePeriod;

    uint64 public genesisTime;

    //Fee's
    uint daoFee;

    //  bool daoflag;
    //  bool payday_f;

    uint DaoCollect;



    // Wallet state
    enum Phase {
        Created,
        Registred,
        Idle,
        Suspected,
        Punished
    }

    Phase public currentPhase;

    //  Events
    event LogPhaseSwitch(Phase newPhase);

    event LogPass(string pass);

    event ToVal(address to, uint val);


    ///@dev constructor
    function HubWallet(address _hubowner, address _dao, whitelist _whitelist, token sharesAddress){
        owner = _hubowner;
        DAO = _dao;
        Whitelist = whitelist(_whitelist);
        Factory = msg.sender;
        genesisTime = uint64(now);

        sharesTokenAddress = token(sharesAddress);

        //1 SNM token is needed to registrate in whitelist
        freezeQuote = 1 * (1 ether / 1 wei);

        // in the future this percent will be defined by factory.
        lockPercent = 30;

        //in promilles
        daoFee = 5;

        freezePeriod = 10 days;

        currentPhase = Phase.Idle;
    }

    function Registration() public returns (bool success){
        if (currentPhase != Phase.Idle) throw;

        if (sharesTokenAddress.balanceOf(this) <= freezeQuote) throw;
        //  LogPass("balance checked okay");
        frozenFunds = freezeQuote;
        frozenTime = uint64(now);
        //Appendix to call register function from Whitelist contract and check it.
        Whitelist.RegisterHub(owner, this, frozenTime);

        currentPhase = Phase.Registred;
        LogPhaseSwitch(currentPhase);

        return true;
    }

    function transfer(address _to, uint _value) public onlyOwner {

        //address _to, uint _value
        //  ToVal(_to, _value);

        if (currentPhase != Phase.Registred) throw;


        uint lockFee = _value * lockPercent / 100;
        uint lock = lockedFunds + lockFee;
        uint limit = lock + frozenFunds;

        uint value = _value - lockFee;

        if (sharesTokenAddress.balanceOf(msg.sender) < (limit + value)) throw;

        lockedFunds = lock;

        sharesTokenAddress.approve(_to, value);
        //  LogPass("done");

    }

    function PayDay() public onlyOwner {

        if (currentPhase != Phase.Registred) throw;
        //  if (daoflag!=true) throw;

        DaoCollect = lockedFunds * daoFee / 1000;
        DaoCollect = DaoCollect + frozenFunds;
        frozenFunds = 0;
        lockedFunds = 0;

        // Comment it for test.
        if (now < (frozenTime + freezePeriod)) throw;

        //For test usage
        //  DaoCollect=0;

        //dao got's 0.5% in such terms.
        sharesTokenAddress.transfer(DAO, DaoCollect);

        Whitelist.UnRegisterHub(owner, this);
        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
        //  payday_f = true;
    }

    function withdraw() public onlyOwner {

        //  if(currentPhase != Phase.Created || currentPhase!=Phase.Idle) throw;
        if (currentPhase != Phase.Idle) throw;
        //  if (daoflag!=true) throw;
        uint amount = sharesTokenAddress.balanceOf(this);

        sharesTokenAddress.transfer(owner, amount);

    }

    /*
      // For test only. Remove.
      function DaoTransfer() public onlyOwner {
      //  if(currentPhase!=Phase.Registred) throw;
    //    if(payday_f!=true) throw;

      //  sharesTokenAddress.transfer(DAO,DaoCollect);

        DaoCollect=0;
        sharesTokenAddress.transfer(DAO,DaoCollect);
        daoflag = true;

      }
      */

    function suspect() public onlyDao {
        if (currentPhase != Phase.Registred) throw;
        //    frozenFunds = 0;
        lockedFunds = sharesTokenAddress.balanceOf(this);
        frozenTime = uint64(now);
        currentPhase = Phase.Suspected;
        LogPhaseSwitch(currentPhase);
        freezePeriod = 120 days;
        Whitelist.UnRegisterHub(owner, this);
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

        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }


}
