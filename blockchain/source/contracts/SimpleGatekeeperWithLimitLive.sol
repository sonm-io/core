pragma solidity ^0.4.23;


import "./SNMMasterchain.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract SimpleGatekeeperWithLimitLive is Ownable {

    using SafeMath for uint256;

    SNMMasterchain token;

    struct Keeper {
        uint256 dayLimit;
        uint256 lastDay;
        uint256 spentToday;
        bool frozen;
    }

    struct TransactionState {
        uint256 commitTS;
        bool paid;
        address keeper;
    }

    mapping(address => Keeper) keepers;

    constructor(address _token) public {
        token = SNMMasterchain(_token);
        owner = msg.sender;
    }

    uint256 public transactionAmount = 0;
    mapping(bytes32 => TransactionState) public paid;

    event PayinTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event CommitTx(address indexed from, uint256 indexed txNumber, uint256 indexed value, uint commitTimestamp);
    event PayoutTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event Suicide(uint block);

    event KeeperChanged(address indexed keeper, uint256 indexed dayLimit);
    event KeeperFrozen(address indexed keeper);
    event KeeperUnfrozen(address indexed keeper);

    function ChangeKeeperLimit(address _keeper, uint256 _limit) public onlyOwner {
        keepers[_keeper].dayLimit = _limit;
        emit KeeperChanged(_keeper, _limit);
    }

    function FreezeKeeper(address _keeper) public {
        // check access of sender
        require(keepers[msg.sender].dayLimit > 0);
        // check that freezing keeper has limit
        require(keepers[_keeper].dayLimit > 0);
        keepers[_keeper].frozen = true;
        emit KeeperFrozen(_keeper);
    }

    function UnFreezeKeeper(address _keeper) public onlyOwner {
        require(keepers[_keeper].dayLimit > 0);
        keepers[_keeper].frozen = false;
        emit KeeperUnfrozen(_keeper);
    }

    function Payin(uint256 _value) public {
        token.transferFrom(msg.sender, this, _value);
        transactionAmount = transactionAmount + 1;
        emit PayinTx(msg.sender, transactionAmount, _value);
    }

    function Payout(address _to, uint256 _value, uint256 _txNumber) public {
        // check that keeper is not frozen
        require(!keepers[msg.sender].frozen);
        require(keepers[msg.sender].dayLimit > 0);

        bytes32 txHash = keccak256(_to, _txNumber, _value);

        // check that transaction is not paid
        require(!paid[txHash].paid);

        if (paid[txHash].commitTS == 0) {
            // check daylimit
            require(underLimit(msg.sender, _value));
            paid[txHash].commitTS = block.timestamp;
            paid[txHash].keeper = msg.sender;
            emit CommitTx(_to, _txNumber, _value, block.timestamp);
        } else {
            require(paid[txHash].keeper == msg.sender);
            require(paid[txHash].commitTS + 1 days >= block.timestamp);
            token.transfer(_to, _value);
            paid[txHash].paid = true;
            emit PayoutTx(_to, _txNumber, _value);
        }
    }

    function underLimit(address _keeper, uint256 _value) internal returns (bool) {
        // reset the spend limit if we're on a different day to last time.
        if (today() > keepers[_keeper].lastDay) {
            keepers[_keeper].spentToday = 0;
            keepers[_keeper].lastDay = today();
        }
        // check to see if there's enough left - if so, subtract and return true.
        // overflow protection                    // dailyLimit check
        if (keepers[_keeper].spentToday + _value >= keepers[_keeper].spentToday &&
        keepers[_keeper].spentToday + _value <= keepers[_keeper].dayLimit) {
            keepers[_keeper].spentToday += _value;
            return true;
        }
        return false;
    }

    function today() private view returns (uint256) {
        // solium-disable-next-line security/no-block-members
        return block.timestamp / 1 days;
    }

    function kill() public onlyOwner {
        uint balance = token.balanceOf(this);
        token.transfer(owner, balance);
        emit Suicide(block.timestamp);
        // solium-disable-line security/no-block-members
        selfdestruct(owner);
    }

}
