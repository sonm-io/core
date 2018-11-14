pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/math/SafeMath.sol";
import "./FixedPoint128.sol";
import "./RatingData.sol";

contract Rating is Ownable {

    using SafeMath for uint256;
    using SafeMath for uint;
    using FixedPoint128 for uint256;

    RatingData ratingData;
    event RatingUpdated(address who);
    event DecayValueUpdated(uint256 decayValue);

    enum Outcome {
        Unclaimed,
        ClaimAccepted,
        ClaimRejected,
        Invalid
    }

    enum Role {
        Worker,
        Master,
        Consumer
    }

    constructor(address _ratingData) public {
        owner = msg.sender;
        ratingData = RatingData(_ratingData);
    }

    function GetRatingData() public view returns (address) {
        return address(ratingData);
    }

    function SetRatingData(address _ratingData) public onlyOwner {
        ratingData = RatingData(_ratingData);
    }

    function TransferData(address to) public onlyOwner {
        ratingData.transferOwnership(to);
    }

    function SetDecayValue(uint256 newDecayValue) public onlyOwner {
        ratingData.SetDecayValue(newDecayValue);
        emit DecayValueUpdated(newDecayValue);
    }

    function DecayValue() public view returns (uint256) {
        return ratingData.DecayValue();
    }

    function Sum(Role role, Outcome outcome, address whose) public view returns (uint256) {
        return ratingData.Sum(uint8(role), uint8(outcome), whose);
    }

    function LastBlockTs(Role role, address who) public view returns (uint256) {
        return ratingData.LastBlockTs(uint8(role), who);
    }


    function Unclaimed(Role role, address whos) public view returns (uint256) {
        uint256 unclaimedSum = ratingData.Sum(uint8(role), uint8(Outcome.Unclaimed), whos);
        uint256 resolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimAccepted), whos);
        uint256 unresolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimRejected), whos);

        uint256 allSum = unclaimedSum.FPAdd(resolvedSum).FPAdd(unresolvedSum);

        if (allSum == 0) {
            return FixedPoint128.FromNatural(1);
        }
        return unclaimedSum.FPDiv(allSum);
    }

    function Unproblematic(Role role, address whos) public view returns (uint256) {
        uint256 unclaimedSum = ratingData.Sum(uint8(role), uint8(Outcome.Unclaimed), whos);
        uint256 resolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimAccepted), whos);
        uint256 unresolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimRejected), whos);

        uint256 allSum = unclaimedSum.FPAdd(resolvedSum).FPAdd(unresolvedSum);

        if (allSum == 0) {
            return FixedPoint128.FromNatural(1);
        }
        return unclaimedSum.FPAdd(resolvedSum).FPDiv(allSum);
    }

    function ClaimResolveRating(Role role, address whos) public view returns (uint256) {
        uint256 resolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimAccepted), whos);
        uint256 unresolvedSum = ratingData.Sum(uint8(role), uint8(Outcome.ClaimRejected), whos);

        uint256 allSum = resolvedSum.FPAdd(unresolvedSum);
        if (allSum == 0) {
            return FixedPoint128.FromNatural(1);
        }
        return resolvedSum.FPDiv(allSum);
    }

    function applyDecay(Role role, address who, uint64 when) private {
        uint64 lastBlockTs = ratingData.LastBlockTs(uint8(role), who);
        require(when >= lastBlockTs);
        ratingData.SetLastBlockTs(uint8(role), who, when);
        if(lastBlockTs == 0) {
            return;
        }
        uint64 timeDiff = when - lastBlockTs;
        uint256 decay = ratingData.DecayValue().FPPow(timeDiff);
        for (uint8 outcome = uint8(Outcome.Unclaimed); outcome < uint8(Outcome.Invalid); outcome++) {
            uint256 sum = ratingData.Sum(uint8(role), outcome, who);
            ratingData.SetSum(uint8(role), outcome, who, sum.FPMul(decay));
        }
    }

    function applyNegativeOutcome(Role role, address who, uint256 cpRating, uint256 fpSum) private {
        uint256 negPart = fpSum.FPMul(cpRating);
        uint256 posPart = fpSum.FPSub(negPart);
        ratingData.IncSum(uint8(role), uint8(Outcome.Unclaimed), who, posPart);
        ratingData.IncSum(uint8(role), uint8(Outcome.ClaimRejected), who, negPart);
    }

    function RegisterOutcome(
        Outcome outcome, address consumer, address supplierWorker, address supplierMaster, uint256 sum, uint64 when
    ) public onlyOwner {
        if(outcome >= Outcome.Invalid){
            revert("invalid outcome specified");
        }
        applyDecay(Role.Consumer, consumer, when);
        applyDecay(Role.Master, supplierMaster, when);
        applyDecay(Role.Worker, supplierWorker, when);
        uint256 fpSum = sum.ToFixedPoint();
        if (outcome == Outcome.Unclaimed || outcome == Outcome.ClaimAccepted) {
            ratingData.IncSum(uint8(Role.Consumer), uint8(outcome), consumer, fpSum);
            ratingData.IncSum(uint8(Role.Worker), uint8(outcome), supplierWorker, fpSum);
            ratingData.IncSum(uint8(Role.Master), uint8(outcome), supplierMaster, fpSum);
        } else {
            uint256 supRating = Unproblematic(Role.Master, supplierMaster);
            applyNegativeOutcome(Role.Consumer, consumer, supRating, fpSum);
            uint256 consumerRating = Unproblematic(Role.Consumer, consumer);
            applyNegativeOutcome(Role.Master, supplierMaster, consumerRating, fpSum);
            applyNegativeOutcome(Role.Worker, supplierWorker, consumerRating, fpSum);
        }
        emit RatingUpdated(consumer);
        emit RatingUpdated(supplierWorker);
        emit RatingUpdated(supplierMaster);
    }
}
