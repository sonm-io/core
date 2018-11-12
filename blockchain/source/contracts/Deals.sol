pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract Deals is Ownable {
    //events

    //enums
    enum DealStatus {
        STATUS_UNKNOWN,
        STATUS_ACCEPTED,
        STATUS_CLOSED
    }
    //data

    struct Deal {
        DealInfo info;
        DealParams params;
    }

    struct DealInfo {
        uint64[] benchmarks;
        address supplierID;
        address consumerID;
        address masterID;
        uint askID;
        uint bidID;
        uint startTime;
    }

    struct DealParams {
        uint duration;
        uint price; //usd * 10^-18
        uint endTime;
        DealStatus status;
        uint blockedBalance;
        uint totalPayout;
        uint lastBillTS;
    }

    mapping(uint => Deal) public deals;

    mapping(address => uint[]) dealsID;

    uint dealsAmount = 0;

    //Constructor
    constructor() public {
        owner = msg.sender;
    }

    function Write(
        uint64[] _benchmarks,
        address _supplierID,
        address _consumerID,
        address _masterID,
        uint _askID,
        uint _bidID,
        uint _duration,
        uint _price,
        uint _startTime,
        uint _endTime,
        DealStatus _status,
        uint _blockedBalance,
        uint _totalPayout,
        uint _lastBillTS) public onlyOwner returns(uint) {

        dealsAmount += 1;

        deals[dealsAmount].info.benchmarks = _benchmarks;
        deals[dealsAmount].info.supplierID = _supplierID;
        deals[dealsAmount].info.consumerID = _consumerID;
        deals[dealsAmount].info.masterID = _masterID;
        deals[dealsAmount].info.askID = _askID;
        deals[dealsAmount].info.bidID = _bidID;
        deals[dealsAmount].info.startTime = _startTime;
        deals[dealsAmount].params.duration = _duration;
        deals[dealsAmount].params.price = _price;
        deals[dealsAmount].params.endTime = _endTime;
        deals[dealsAmount].params.status = _status;
        deals[dealsAmount].params.blockedBalance = _blockedBalance;
        deals[dealsAmount].params.totalPayout = _totalPayout;
        deals[dealsAmount].params.lastBillTS = _lastBillTS;


        return dealsAmount;
    }

    function SetDealBenchmarks(uint dealID, uint64[] _benchmarks) public onlyOwner {
        deals[dealID].info.benchmarks = _benchmarks;
    }

    function SetDealSupplierID(uint dealID, address _supplierID) public onlyOwner {
        deals[dealID].info.supplierID = _supplierID;
    }

    function SetDealConsumerID(uint dealID, address _consumerID) public onlyOwner {
        deals[dealID].info.consumerID = _consumerID;
    }

    function SetDealMasterID(uint dealID, address _masterID) public onlyOwner {
        deals[dealID].info.masterID = _masterID;
    }

    function SetDealAskID(uint dealID, uint _askID) public onlyOwner {
        deals[dealID].info.askID = _askID;
    }

    function SetDealBidID(uint dealID, uint _bidID) public onlyOwner {
        deals[dealID].info.bidID = _bidID;
    }

    function SetDealStartTime(uint dealID, uint _startTime) public onlyOwner {
        deals[dealID].info.startTime = _startTime;
    }

    function SetDealStatus(uint dealID, DealStatus _status) public onlyOwner {
        deals[dealID].params.status = _status;
    }

    function SetDealEndTime(uint dealID, uint _endTime) public onlyOwner {
        deals[dealID].params.endTime = _endTime;
    }

    function SetDealBlockedBalance(uint dealID, uint _blockedBalance) public onlyOwner {
        deals[dealID].params.blockedBalance = _blockedBalance;
    }

    function SetDealLastBillTS(uint dealID, uint _lastBillTS) public onlyOwner {
        deals[dealID].params.lastBillTS = _lastBillTS;
    }

    function SetDealTotalPayout(uint dealID, uint _totalPayout) public onlyOwner {
        deals[dealID].params.totalPayout = _totalPayout;
    }

    function SetDealPrice(uint dealID, uint _price) public onlyOwner {
        deals[dealID].params.price = _price;
    }

    function SetDealDuration(uint dealID, uint _duration) public onlyOwner {
        deals[dealID].params.duration = _duration;
    }

    function IncreaseDealsAmount() public onlyOwner {
        dealsAmount += 1;
    }


    //getters

    function GetDealInfo(uint dealID) public view
    returns (
        uint64[] benchmarks,
        address supplierID,
        address consumerID,
        address masterID,
        uint askID,
        uint bidID,
        uint startTime
    ) {
        DealInfo memory info = deals[dealID].info;
        return (
        info.benchmarks,
        info.supplierID,
        info.consumerID,
        info.masterID,
        info.askID,
        info.bidID,
        info.startTime
        );
    }

    function GetDealParams(uint dealID) public view
    returns (
        uint duration,
        uint price,
        uint endTime,
        DealStatus status,
        uint blockedBalance,
        uint totalPayout,
        uint lastBillTS
    ) {
        DealParams memory params = deals[dealID].params;
        return (
        params.duration,
        params.price,
        params.endTime,
        params.status,
        params.blockedBalance,
        params.totalPayout,
        params.lastBillTS
        );
    }

    function GetDealBenchmarks(uint dealID) public view returns(uint64[]) {
        return deals[dealID].info.benchmarks;
    }

    function GetDealConsumerID(uint dealID) public view returns(address) {
        return deals[dealID].info.consumerID;
    }

    function GetDealSupplierID(uint dealID) public view returns(address) {
        return deals[dealID].info.supplierID;
    }

    function GetDealMasterID(uint dealID) public view returns(address) {
        return deals[dealID].info.masterID;
    }

    function GetDealAskID(uint dealID) public view returns(uint) {
        return deals[dealID].info.askID;
    }

    function GetDealBidID(uint dealID) public view returns(uint) {
        return deals[dealID].info.bidID;
    }

    function GetDealStartTime(uint dealID) public view returns(uint) {
        return deals[dealID].info.startTime;
    }

    function GetDealDuration(uint dealID) public view returns(uint) {
        return deals[dealID].params.duration;
    }

    function GetDealPrice(uint dealID) public view returns(uint) {
        return deals[dealID].params.price;
    }

    function GetDealEndTime(uint dealID) public view returns(uint) {
        return deals[dealID].params.endTime;
    }

    function GetDealStatus(uint dealID) public view returns(DealStatus) {
        return deals[dealID].params.status;
    }

    function GetDealBlockedBalance(uint dealID) public view returns(uint) {
        return deals[dealID].params.blockedBalance;
    }

    function GetDealTotalPayout(uint dealID) public view returns(uint) {
        return deals[dealID].params.totalPayout;
    }

    function GetDealLastBillTS(uint dealID) public view returns(uint) {
        return deals[dealID].params.lastBillTS;
    }

    function GetDealsAmount() public view returns(uint) {
        return dealsAmount;
    }

}
