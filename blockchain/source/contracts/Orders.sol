pragma solidity ^0.4.23;

import "./Administratable.sol";
import "./ProfileRegistry.sol";

contract Orders is Administratable {
    //events

    //enums

    enum OrderStatus {
        UNKNOWN,
        ORDER_INACTIVE,
        ORDER_ACTIVE
    }

    enum OrderType {
        ORDER_UNKNOWN,
        ORDER_BID,
        ORDER_ASK
    }

    //DATA
    struct Order {
        OrderInfo info;
        OrderParams params;
    }

    struct OrderInfo {
        OrderType orderType;
        address author;
        address counterparty;
        uint duration;
        uint price;
        bool[] netflags;
        ProfileRegistry.IdentityLevel identityLevel;
        address blacklist;
        bytes32 tag;
        uint64[] benchmarks;
        uint frozenSum;
        uint requiredRating;
    }

    struct OrderParams {
        OrderStatus orderStatus;
        uint dealID;
    }

    mapping(uint => Order) orders;

    uint ordersAmount = 0;

    constructor() public {
        owner = msg.sender;
        administrator = msg.sender;
    }

    // Accessors, listed in the same order as the appear in `Order` struct
    function GetOrderType(uint orderID) public view returns (OrderType) {
        return orders[orderID].info.orderType;
    }

    function SetOrderType(uint orderID, OrderType newType) public onlyOwner {
        orders[orderID].info.orderType = newType;
    }

    function GetOrderAuthor(uint orderID) public view returns (address) {
        return orders[orderID].info.author;
    }

    function SetOrderAuthor(uint orderID, address newAuthor) public onlyOwner {
        orders[orderID].info.author = newAuthor;
    }

    function GetOrderCounterparty(uint orderID) public view returns (address) {
        return orders[orderID].info.counterparty;
    }

    function SetOrderCounterparty(uint orderID, address newCounterparty) public onlyOwner {
        orders[orderID].info.counterparty = newCounterparty;
    }

    function GetOrderDuration(uint orderID) public view returns (uint) {
        return orders[orderID].info.duration;
    }

    function SetOrderDuration(uint orderID, uint newDuration) public onlyOwner {
        orders[orderID].info.duration = newDuration;
    }

    function GetOrderPrice(uint orderID) public view returns (uint) {
        return orders[orderID].info.price;
    }

    function SetOrderPrice(uint orderID, uint newPrice) public onlyOwner {
        orders[orderID].info.price = newPrice;
    }

    function GetOrderNetflags(uint orderID) public view returns (bool[]) {
        return orders[orderID].info.netflags;
    }

    function SetOrderNetflags(uint orderID, bool[] _netflags) public onlyOwner {
        orders[orderID].info.netflags = _netflags;
    }

    function GetOrderIdentityLevel(uint orderID) public view returns (ProfileRegistry.IdentityLevel) {
        return orders[orderID].info.identityLevel;
    }

    function SetOrderIdentityLevel(uint orderID, ProfileRegistry.IdentityLevel newLevel) public onlyOwner {
        orders[orderID].info.identityLevel = newLevel;
    }

    function GetOrderBlacklist(uint orderID) public view returns (address) {
        return orders[orderID].info.blacklist;
    }

    function SetOrderBlacklist(uint orderID, address newBlacklist) public onlyOwner {
        orders[orderID].info.blacklist = newBlacklist;
    }

    function GetOrderTag(uint orderID) public view returns (bytes32) {
        return orders[orderID].info.tag;
    }

    function SetOrderTag(uint orderID, bytes32 _tag) public onlyOwner {
        orders[orderID].info.tag = _tag;
    }

    function GetOrderBenchmarks(uint orderID) public view returns (uint64[]) {
        return orders[orderID].info.benchmarks;
    }

    function SetOrderBenchmarks(uint orderID, uint64[] _benchmarks) public onlyOwner {
        orders[orderID].info.benchmarks = _benchmarks;
    }

    function GetOrderFrozenSum(uint orderID) public view returns (uint) {
        return orders[orderID].info.frozenSum;
    }

    function SetOrderFrozenSum(uint orderID, uint newFrozenSum) public onlyOwner {
        orders[orderID].info.frozenSum = newFrozenSum;
    }

    function GetOrderRequiredRating(uint orderID) public view returns (uint) {
        return orders[orderID].info.requiredRating;
    }

    function SetOrderRequiredRating(uint orderID, uint newRequiredRating) public onlyOwner {
        orders[orderID].info.requiredRating = newRequiredRating;
    }

    function GetOrderStatus(uint orderID) public view returns (OrderStatus) {
        return orders[orderID].params.orderStatus;
    }

    function SetOrderStatus(uint orderID, OrderStatus _status) public onlyOwner {
        orders[orderID].params.orderStatus = _status;
    }

    function GetOrderDealID(uint orderID) public view returns (uint) {
        return orders[orderID].params.dealID;
    }

    function SetOrderDealID(uint orderID, uint _dealID) public onlyOwner {
        orders[orderID].params.dealID = _dealID;
    }

    // Cummulative Order actions. Those are in fact only helpers.
    function Write(
        OrderType _orderType,
        OrderStatus _orderStatus,
        address _author,
        address _counterparty,
        uint _duration,
        uint256 _price,
        bool[] _netflags,
        ProfileRegistry.IdentityLevel _identityLevel,
        address _blacklist,
        bytes32 _tag,
        uint64[] _benchmarks,
        uint _frozenSum,
        uint _dealID) public onlyOwner  returns(uint) {

        ordersAmount += 1;

        orders[ordersAmount].info.orderType = _orderType;
        orders[ordersAmount].info.author = _author;
        orders[ordersAmount].info.counterparty = _counterparty;
        orders[ordersAmount].info.duration = _duration;
        orders[ordersAmount].info.price = _price;
        orders[ordersAmount].info.netflags = _netflags;
        orders[ordersAmount].info.identityLevel = _identityLevel;
        orders[ordersAmount].info.blacklist = _blacklist;
        orders[ordersAmount].info.tag = _tag;
        orders[ordersAmount].info.benchmarks = _benchmarks;
        orders[ordersAmount].info.frozenSum = _frozenSum;
        orders[ordersAmount].params.orderStatus = _orderStatus;
        orders[ordersAmount].params.dealID = _dealID;

        return ordersAmount;
    }

    function DeleteOrder(uint orderID) public onlyOwner {
        delete orders[orderID];
    }

    function GetOrderInfo(uint orderID) public view
    returns (
        OrderType orderType,
        address author,
        address counterparty,
        uint duration,
        uint price,
        bool[] netflags,
        ProfileRegistry.IdentityLevel identityLevel,
        address blacklist,
        bytes32 tag,
        uint64[] benchmarks,
        uint frozenSum,
        uint requiredRating
    ){
        OrderInfo memory info = orders[orderID].info;
        return (
        info.orderType,
        info.author,
        info.counterparty,
        info.duration,
        info.price,
        info.netflags,
        info.identityLevel,
        info.blacklist,
        info.tag,
        info.benchmarks,
        info.frozenSum,
        info.requiredRating
        );
    }

    function GetOrderParams(uint orderID) public view
    returns (
        OrderStatus orderStatus,
        uint dealID
    ){
        OrderParams memory params = orders[orderID].params;
        return (
        params.orderStatus,
        params.dealID
        );
    }

    // ordersAmount accessors. Generally setter should not be used, but let it be here just in case.
    function GetOrdersAmount() public view returns (uint) {
        return ordersAmount;
    }

    function SetOrdersAmount(uint newOrdersAmount) public onlyOwner {
        ordersAmount = newOrdersAmount;
    }
}
