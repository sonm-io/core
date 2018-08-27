pragma solidity ^0.4.23;
import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/math/SafeMath.sol";


contract RatingData is Ownable {
    mapping(address => uint256) positiveSum;
    mapping(address => uint256) negativeSum;
    mapping(address => uint64) lastBlock;

    uint256 decayEpoch;
    uint256 decayValue;
    uint256 precision;

    enum Outcome {
        Positive,
        Negative
    }
}

contract Rating is Ownable {

    using SafeMath for uint256;

    event RatingUpdated(address who, uint price);
    event DecayEpochUpdated(uint256 decayEpoch);
    event DecayValueUpdated(uint256 decayValue);



    constructor(uint256 _decayEpoch, uint256 _decayValue, uint256 _precision) public {
//        owner = msg.sender;
//        decayEpoch = _decayEpoch;
//        decayValue = _decayValue;
//        precision = _precision;
    }

    function SetDecayEpoch(uint256 newDecayEpoch) onlyOwner {
    }

    function SetDecayValue(uint256 newDecayEpoch) onlyOwner {
    }

    function DecayValue() public view returns (uint256) {

    }

    function DecayEpoch() public view returns (uint256) {

    }

    function Precision() public view returns (uint256) {

    }

    function Current(address whom) public view returns (uint256) {
        return 0;
//        RatingData memory rating = ratings[whom];
//        return rating.positiveSum.Mul(ratingPrecision).Div(rating.allSum);
    }



    function RegisterOutcome(RatingData.Outcome outcome, address consumer, address supplier, uint256 sum) onlyOwner {
//        RatingData memory consumerRating = ratings[consumer];
//        RatingData memory supplierRating = ratings[supplier];
//        if (outcome == Outcome.Negative) {
//            decay = CalculateDecay(consumerRating.lastBlock, block.Timestamp);
//        }
    }

    function CalculateDecay(uint256 from, uint256 to) public pure returns(uint) {
        return 1;
    }



}
