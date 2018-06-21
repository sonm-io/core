pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/token/ERC20/StandardToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract SimpleGatekeeperWithLimit is Ownable {

    using SafeMath for uint256;

    StandardToken token;

    struct Keeper {
        uint256 dayLimit;
        uint256 lastDay;
        uint256 spentToday;
        bool frozen;
    }

    mapping(address => Keeper) keepers;

    constructor(address _token) public {
        token = StandardToken(_token);
        owner = msg.sender;
    }

    uint256 public transactionAmount = 0;
    mapping(bytes32 => bool) public paid;

    event PayInTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event PayoutTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event Suicide(uint block);

    event KeeperChanged(address keeper, uint256 dayLimit);
    event KeeperFrozen(address keeper);
    event KeeperUnfrozen(address keeper);

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
        require(keepers[msg.sender].dayLimit > 0);
        keepers[_keeper].frozen = false;
        emit KeeperUnfrozen(_keeper);
    }

    function PayIn(uint256 _value) public {
        require(token.transferFrom(msg.sender, this, _value));
        transactionAmount = transactionAmount + 1;
        emit PayInTx(msg.sender, transactionAmount, _value);
    }

    function Payout(address _to, uint256 _value, uint256 _txNumber) public {
        // check that keeper not frozen
        require(!keepers[msg.sender].frozen);
        // check day limit
        require(underLimit(msg.sender, _value));

        bytes32 txHash = keccak256(_to, _txNumber, _value);
        require(!paid[txHash]);
        require(token.transfer(_to, _value));
        paid[txHash] = true;
        emit PayoutTx(_to, _txNumber, _value);
    }

    function kill() public onlyOwner {
        uint balance = token.balanceOf(this);
        require(token.transfer(owner, balance));
        emit Suicide(block.timestamp);
        // solium-disable-line security/no-block-members
        selfdestruct(owner);
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

}
