pragma solidity ^0.4.4;


//Raw prototype for Hub wallet contract.


import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/token/ERC20.sol";
import "./Whitelist.sol";


contract HubWallet is Ownable {

    modifier onlyDao() { require(msg.sender == DAO); _; }

    address public DAO;

    address public Factory;

    Whitelist whitelist;

    ERC20 public sharesTokenAddress;

    // FreezeQuote - it is defined amount of tokens need to be frozen on  this contract.
    uint public freezeQuote;

    uint public frozenFunds;

    //lockedFunds in percentage, which will be locked for every payday period.
    uint public lockPercent;

    uint public lockedFunds = 0;

    //TIMELOCK
    uint64 public frozenTime;

    uint public freezePeriod;

    uint64 public genesisTime;

    //Fees
    uint daoFee;

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
    event ToVal(address to, uint val);


    ///@dev constructor
    function HubWallet(address _hubowner, address _dao, Whitelist _whitelist, ERC20 sharesAddress){
        owner = _hubowner;
        DAO = _dao;
        whitelist = _whitelist;
        Factory = msg.sender;
        genesisTime = uint64(now);
        sharesTokenAddress = sharesAddress;

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
        require(currentPhase == Phase.Idle);
        require(sharesTokenAddress.balanceOf(this) > freezeQuote);

        frozenFunds = freezeQuote;
        frozenTime = uint64(now);
        //Appendix to call register function from Whitelist contract and check it.
        whitelist.RegisterHub(this, frozenTime);

        currentPhase = Phase.Registred;
        LogPhaseSwitch(currentPhase);

        return true;
    }

    function transfer(address _to, uint _value) public onlyOwner {
        require(currentPhase == Phase.Registred);

        uint lockFee = _value * lockPercent / 100;
        uint lock = lockedFunds + lockFee;
        uint limit = lock + frozenFunds;

        uint value = _value - lockFee;

        require(sharesTokenAddress.balanceOf(msg.sender) >= limit + value);

        lockedFunds = lock;

        sharesTokenAddress.approve(_to, value);
    }

    function PayDay() public onlyOwner {
        require(currentPhase == Phase.Registred);

        DaoCollect = lockedFunds * daoFee / 1000;
        DaoCollect = DaoCollect + frozenFunds;
        frozenFunds = 0;
        lockedFunds = 0;

        require(now < frozenTime + freezePeriod);

        //dao got's 0.5% in such terms.
        sharesTokenAddress.transfer(DAO, DaoCollect);

        whitelist.UnRegisterHub(this);
        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }

    function withdraw() public onlyOwner {
        require(currentPhase == Phase.Idle);
        uint amount = sharesTokenAddress.balanceOf(this);
        sharesTokenAddress.transfer(owner, amount);
    }

    function suspect() public onlyDao {
        require(currentPhase == Phase.Registred);

        lockedFunds = sharesTokenAddress.balanceOf(this);
        frozenTime = uint64(now);
        currentPhase = Phase.Suspected;
        LogPhaseSwitch(currentPhase);
        freezePeriod = 120 days;
        whitelist.UnRegisterHub(this);
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

        currentPhase = Phase.Idle;
        LogPhaseSwitch(currentPhase);
    }
}
