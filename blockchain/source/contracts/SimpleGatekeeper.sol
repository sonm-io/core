pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/token/ERC20/StandardToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract SimpleGatekeeper is Ownable {

    using SafeMath for uint256;

    StandardToken token;

    constructor(address _token) public {
        token = StandardToken(_token);
        owner = msg.sender;
    }

    uint256 public transactionAmount = 0;
    mapping(bytes32 => bool) public paid;

    event PayInTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event PayoutTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);
    event Suicide(uint block);

    function PayIn(uint256 _value) public {
        require(token.transferFrom(msg.sender, this, _value));
        transactionAmount = transactionAmount + 1;
        emit PayInTx(msg.sender, transactionAmount, _value);
    }

    function Payout(address _to, uint256 _value, uint256 _txNumber) public onlyOwner{
        bytes32 txHash = keccak256(_to, _txNumber, _value);
        require(!paid[txHash]);
        require(token.transfer(_to, _value));
        paid[txHash] = true;
        emit PayoutTx(_to, _txNumber, _value);
    }

    function kill() public onlyOwner{
        uint balance = token.balanceOf(this);
        require(token.transfer(owner, balance));
        emit Suicide(block.timestamp); // solium-disable-line security/no-block-members
        selfdestruct(owner);
    }

}
