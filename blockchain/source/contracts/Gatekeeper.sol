pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/token/ERC20/StandardToken.sol";
import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract Gatekeeper is Ownable {

    using SafeMath for uint256;

    constructor(address _token) public{
        token = StandardToken(_token);
        owner = msg.sender;
    }

    StandardToken token;

    uint256 transactionCount = 0;

    uint256 requiredTransactionAmountForBlock = 512;

    mapping(uint256 => bytes32) roots;

    uint256 rootsCount = 0;

    mapping(bytes32 => mapping(bytes32 => bool)) paidTx;

    event PayInTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);

    event PayoutTx(address indexed from, uint256 indexed txNumber, uint256 indexed value);

    event BlockEmitted(uint256 indexed blockNumber);

    event TransactionAmountForBlockChanged(uint256 indexed amount);

    event RootAdded(uint256 indexed id, bytes32 indexed root);

    function SetTransactionAmountForBlock(uint256 _amount) onlyOwner public {
        require(_amount > transactionCount, "new amount lower than current counter");
        requiredTransactionAmountForBlock = _amount;
        emit TransactionAmountForBlockChanged(_amount);
    }

    function PayIn(uint256 _value) public {
        require(token.transferFrom(msg.sender, this, _value));
        transactionCount = transactionCount + 1;

        emit PayInTx(msg.sender, transactionCount, _value);

        if (transactionCount == requiredTransactionAmountForBlock) {
            transactionCount = 0;
            emit BlockEmitted(block.number);
        }
    }

    function GetCurrentTransactionAmountForBlock() view public returns (uint256){
        return requiredTransactionAmountForBlock;
    }

    function GetTransactionCount() view public returns (uint256){
        return transactionCount;
    }

    function VoteRoot(bytes32 _root) public onlyOwner {
        rootsCount = rootsCount + 1;
        roots[rootsCount] = _root;
        emit RootAdded(rootsCount, _root);
    }

    function GetRootsCount() view public returns (uint256){
        return rootsCount;
    }

    function Payout(bytes _proof, uint _root, address _from, uint256 _txNumber, uint256 _value) public {
        bytes32 txHash = keccak256(_from, _txNumber, _value);
        require(!paidTx[roots[_root]][txHash]);
        require(verifyProof(_proof, roots[_root], txHash));
        require(token.transfer(_from, _value));
        paidTx[roots[_root]][txHash] = true;
        emit PayoutTx(_from, _txNumber, _value);
    }

    function verifyProof(bytes _proof, bytes32 _root, bytes32 _leaf) pure internal returns (bool) {
        if (_proof.length % 32 != 0) {
            return false;
        }

        bytes32 proofElement;
        bytes32 computedHash = _leaf;

        for (uint16 i = 32; i <= _proof.length; i += 32) {
            assembly {// solium-disable-line security/no-inline-assembly
                proofElement := mload(add(_proof, i)) // solium-disable-line security/no-inline-assembly
            }

            if (computedHash < proofElement) {
                computedHash = keccak256(computedHash, proofElement);
            } else {
                computedHash = keccak256(proofElement, computedHash);
            }
        }

        return computedHash == _root;
    }
}
