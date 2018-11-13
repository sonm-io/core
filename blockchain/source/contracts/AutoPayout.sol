pragma solidity ^0.4.23;

import "./SNM.sol";
import "./SimpleGatekeeperWithLimit.sol";

contract AutoPayout is Ownable {

    using SafeMath for uint256;

    event AutoPayoutChanged(address indexed master, address indexed target, uint256 indexed limit);
    event AutoPayout(address indexed master);
    event Suicide(uint block);

    struct PayoutSetting {
        uint256 lowLimit;
        address target;
    }

    SNM token;

    SimpleGatekeeperWithLimit gatekeeper;

    mapping(address => PayoutSetting) public allowedPayouts;

    constructor(address _token, address _gatekeeper) public {
        token = SNM(_token);
        gatekeeper = SimpleGatekeeperWithLimit(_gatekeeper);
        owner = msg.sender;
    }

    function SetAutoPayout(uint256 _limit, address _target) public {
        allowedPayouts[msg.sender].lowLimit = _limit;
        allowedPayouts[msg.sender].target = _target;
        emit AutoPayoutChanged(msg.sender, _target, _limit);
    }

    function DoAutoPayout(address _master) public {
        uint256 balance = token.balanceOf(_master);
        require(balance >= allowedPayouts[_master].lowLimit);

        token.transferFrom(_master, address(this), balance);

        token.approve(gatekeeper, balance);
        gatekeeper.PayinTargeted(balance, allowedPayouts[_master].target);
        emit AutoPayout(_master);
    }

    function kill() public onlyOwner {
        require(token.transfer(owner, token.balanceOf(address(this))));
        emit Suicide(block.timestamp);
        // solium-disable-line security/no-block-members
        selfdestruct(owner);
    }
}
