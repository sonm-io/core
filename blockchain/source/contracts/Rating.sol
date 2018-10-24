pragma solidity ^0.4.23;
import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/math/SafeMath.sol";
import "./FixedPoint.sol";
import "./transferable.sol";


contract RatingData is CreatorOwnable {
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

    function SetDecayEpoch(uint256 newDecayEpoch) onlyOwner {
        decayEpoch = newDecayEpoch;
    }

    function SetDecayValue(uint256 newDecayValue) onlyOwner {
        decayValue = newDecayValue;
    }

    function DecayValue() public view returns (uint256) {
        return decayValue;
    }

    function DecayEpoch() public view returns (uint256) {
        return decayEpoch;
    }
}

contract Rating is Ownable {

    using SafeMath for uint256;

    RatingData ratingData;
    FixedPoint fp;
    event RatingUpdated(address who, uint price);
    event DecayEpochUpdated(uint256 decayEpoch);
    event DecayValueUpdated(uint256 decayValue);



    constructor(address _ratingData, address _fp) public {
        owner = msg.sender;
        fp = FixedPoint(_fp);
        ratingData = RatingData(_ratingData);
    }

    function transferData(address to) onlyOwner {
        ratingData.transferOwnership(to);
    }

    function SetDecayEpoch(uint256 newDecayEpoch) onlyOwner {
        ratingData.SetDecayEpoch(newDecayEpoch);
    }

    function SetDecayValue(uint256 newDecayValue) onlyOwner {
        ratingData.SetDecayValue(newDecayValue);
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
