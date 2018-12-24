pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/lifecycle/Pausable.sol";
import "./SNM.sol";
import "./Blacklist.sol";
import "./OracleUSD.sol";
import "./ProfileRegistry.sol";
import "./Administratum.sol";
import "./Orders.sol";
import "./Deals.sol";
import "./ChangeRequests.sol";


contract Market is Ownable, Pausable {

    using SafeMath for uint256;

    // DECLARATIONS

    enum BlacklistPerson {
        BLACKLIST_NOBODY,
        BLACKLIST_WORKER,
        BLACKLIST_MASTER
    }


    // EVENTS
    event OrderPlaced(uint indexed orderID);
    event OrderUpdated(uint indexed orderID);

    event DealOpened(uint indexed dealID);
    event DealUpdated(uint indexed dealID);

    event Billed(uint indexed dealID, uint indexed paidAmount);

    event DealChangeRequestSet(uint indexed changeRequestID);
    event DealChangeRequestUpdated(uint indexed changeRequestID);

    event WorkerAnnounced(address indexed worker, address indexed master);
    event WorkerConfirmed(address indexed worker, address indexed master);
    event WorkerRemoved(address indexed worker, address indexed master);

    event NumBenchmarksUpdated(uint indexed newNum);
    event NumNetflagsUpdated(uint indexed newNum);

    // VARS

    uint constant MAX_BENCHMARKS_VALUE = 2 ** 63;

    SNM token;

    Blacklist bl;

    OracleUSD oracle;

    ProfileRegistry profiles;

    Orders ordersCrud;

    Administratum administratum;

    Deals dealsCrud;

    ChangeRequests changeRequestsCrud;

    uint requestsAmount = 0;

    // current length of benchmarks
    uint benchmarksQuantity;

    // current length of netflags
    uint netflagsQuantity;

    // INIT

    constructor(address _token,
        address _blacklist,
        address _oracle,
        address _profileRegistry,
        address _administratum,
        address _orders,
        address _deals,
        address _changeRequests,
        uint _benchmarksQuantity,
        uint _netflagsQuantity
        ) public {
        token = SNM(_token);
        bl = Blacklist(_blacklist);
        oracle = OracleUSD(_oracle);
        profiles = ProfileRegistry(_profileRegistry);
        administratum = Administratum(_administratum);
        ordersCrud = Orders(_orders);
        dealsCrud = Deals(_deals);
        changeRequestsCrud = ChangeRequests(_changeRequests);
        benchmarksQuantity = _benchmarksQuantity;
        netflagsQuantity = _netflagsQuantity;
    }

    // EXTERNAL

    // Order functions

    function PlaceOrder(
        Orders.OrderType _orderType,
        address _counterpartyID,
        uint _duration,
        uint _price,
        bool[] _netflags,
        ProfileRegistry.IdentityLevel _identityLevel,
        address _blacklist,
        bytes32 _tag,
        uint64[] _benchmarks
    ) whenNotPaused public returns (uint) {

        require(_identityLevel >= ProfileRegistry.IdentityLevel.ANONYMOUS);
        require(_netflags.length <= netflagsQuantity);
        require(_benchmarks.length <= benchmarksQuantity);

        for (uint i = 0; i < _benchmarks.length; i++) {
            require(_benchmarks[i] < MAX_BENCHMARKS_VALUE);
        }

        uint lockedSum = 0;

        if (_orderType == Orders.OrderType.ORDER_BID) {
            if (_duration == 0) {
                lockedSum = CalculatePayment(_price, 1 hours);
            } else if (_duration < 1 days) {
                lockedSum = CalculatePayment(_price, _duration);
            } else {
                lockedSum = CalculatePayment(_price, 1 days);
            }
            require(token.transferFrom(msg.sender, address(this), lockedSum));
        }


        ordersCrud.Write(
            _orderType,
            Orders.OrderStatus.ORDER_ACTIVE,
            msg.sender,
            _counterpartyID,
            _duration,
            _price,
            _netflags,
            _identityLevel,
            _blacklist,
            _tag,
            _benchmarks,
            lockedSum,
            0
        );

        emit OrderPlaced(ordersCrud.GetOrdersAmount());

        return ordersCrud.GetOrdersAmount();
    }

    function CancelOrder(uint orderID) public returns (bool) {
        require(orderID <= ordersCrud.GetOrdersAmount());

        address author = ordersCrud.GetOrderAuthor(orderID);

        uint frozenSum = ordersCrud.GetOrderFrozenSum(orderID);

        Orders.OrderStatus orderStatus = ordersCrud.GetOrderStatus(orderID);


        require(orderStatus == Orders.OrderStatus.ORDER_ACTIVE);
        require(author == msg.sender || administratum.GetMaster(author) == msg.sender);

        require(token.transfer(msg.sender, frozenSum));

        ordersCrud.SetOrderStatus(orderID, Orders.OrderStatus.ORDER_INACTIVE);

        emit OrderUpdated(orderID);

        return true;
    }

    function QuickBuy(uint askID, uint buyoutDuration) public whenNotPaused {

        require(ordersCrud.GetOrderType(askID) == Orders.OrderType.ORDER_ASK);
        require(ordersCrud.GetOrderStatus(askID) == Orders.OrderStatus.ORDER_ACTIVE);

        require(ordersCrud.GetOrderDuration(askID) >= buyoutDuration);
        require(profiles.GetProfileLevel(msg.sender) >= ordersCrud.GetOrderIdentityLevel(askID));
        require(bl.Check(ordersCrud.GetOrderBlacklist(askID), msg.sender) == false);
        require(
            bl.Check(msg.sender, administratum.GetMaster(ordersCrud.GetOrderAuthor(askID))) == false
            && bl.Check(ordersCrud.GetOrderAuthor(askID), msg.sender) == false);

        PlaceOrder(
            Orders.OrderType.ORDER_BID,
            administratum.GetMaster(ordersCrud.GetOrderAuthor(askID)),
            buyoutDuration,
            ordersCrud.GetOrderPrice(askID),
            ordersCrud.GetOrderNetflags(askID),
            ProfileRegistry.IdentityLevel.ANONYMOUS,
            address(0),
            bytes32(0),
            ordersCrud.GetOrderBenchmarks(askID));

        OpenDeal(askID, ordersCrud.GetOrdersAmount());
    }

    // Deal functions

    function OpenDeal(uint _askID, uint _bidID) whenNotPaused public {
        require(
            ordersCrud.GetOrderStatus(_askID) == Orders.OrderStatus.ORDER_ACTIVE
            && ordersCrud.GetOrderStatus(_bidID) == Orders.OrderStatus.ORDER_ACTIVE);
        require(ordersCrud.GetOrderCounterparty(_askID) == 0x0 || ordersCrud.GetOrderCounterparty(_askID) == ordersCrud.GetOrderAuthor(_bidID));
        require(
            ordersCrud.GetOrderCounterparty(_bidID) == 0x0
            || ordersCrud.GetOrderCounterparty(_bidID) == administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID))
            || ordersCrud.GetOrderCounterparty(_bidID) == ordersCrud.GetOrderAuthor(_askID));
        require(ordersCrud.GetOrderType(_askID) == Orders.OrderType.ORDER_ASK);
        require(ordersCrud.GetOrderType(_bidID) == Orders.OrderType.ORDER_BID);
        require(!bl.Check(ordersCrud.GetOrderBlacklist(_bidID), administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID))));
        require(!bl.Check(ordersCrud.GetOrderBlacklist(_bidID), ordersCrud.GetOrderAuthor(_askID)));
        require(!bl.Check(ordersCrud.GetOrderAuthor(_bidID), administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID))));
        require(!bl.Check(ordersCrud.GetOrderAuthor(_bidID), ordersCrud.GetOrderAuthor(_askID)));
        require(!bl.Check(ordersCrud.GetOrderBlacklist(_askID), ordersCrud.GetOrderAuthor(_bidID)));
        require(!bl.Check(administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID)), ordersCrud.GetOrderAuthor(_bidID)));
        require(!bl.Check(ordersCrud.GetOrderAuthor(_askID), ordersCrud.GetOrderAuthor(_bidID)));
        require(ordersCrud.GetOrderPrice(_askID) <= ordersCrud.GetOrderPrice(_bidID));
        require(ordersCrud.GetOrderDuration(_askID) >= ordersCrud.GetOrderDuration(_bidID));
        // profile level check
        require(profiles.GetProfileLevel(ordersCrud.GetOrderAuthor(_bidID)) >= ordersCrud.GetOrderIdentityLevel(_askID));
        require(profiles.GetProfileLevel(administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID))) >= ordersCrud.GetOrderIdentityLevel(_bidID)); //bug

        bool[] memory askNetflags = ordersCrud.GetOrderNetflags(_askID);
        if (askNetflags.length < netflagsQuantity) {
            askNetflags = ResizeNetflags(askNetflags);
        }

        bool[] memory bidNetflags = ordersCrud.GetOrderNetflags(_bidID);
        if (bidNetflags.length < netflagsQuantity) {
            bidNetflags = ResizeNetflags(bidNetflags);
        }

        for (uint i = 0; i < netflagsQuantity; i++) {
            // implementation: when bid contains requirement, ask necessary needs to have this
            // if ask have this one - pass
            require(!bidNetflags[i] || askNetflags[i]);
        }

        uint64[] memory askBenchmarks = ordersCrud.GetOrderBenchmarks(_askID);
        if (askBenchmarks.length < benchmarksQuantity) {
            askBenchmarks = ResizeBenchmarks(askBenchmarks);
        }

        uint64[] memory bidBenchmarks = ordersCrud.GetOrderBenchmarks(_bidID);
        if (bidBenchmarks.length < benchmarksQuantity) {
            bidBenchmarks = ResizeBenchmarks(bidBenchmarks);
        }

        for (i = 0; i < benchmarksQuantity; i++) {
            require(askBenchmarks[i] >= bidBenchmarks[i]);
        }

        ordersCrud.SetOrderStatus(_askID, Orders.OrderStatus.ORDER_INACTIVE);
        ordersCrud.SetOrderStatus(_bidID, Orders.OrderStatus.ORDER_INACTIVE);

        emit OrderUpdated(_askID);
        emit OrderUpdated(_bidID);


        // if deal is normal
        if (ordersCrud.GetOrderDuration(_askID) != 0) {
            uint endTime = block.timestamp.add(ordersCrud.GetOrderDuration(_bidID));
        } else {
            endTime = 0;
            // `0` - for spot deal
        }

        dealsCrud.IncreaseDealsAmount();

        uint dealID = dealsCrud.GetDealsAmount();

        dealsCrud.SetDealBenchmarks(dealID, askBenchmarks);
        dealsCrud.SetDealConsumerID(dealID, ordersCrud.GetOrderAuthor(_bidID));
        dealsCrud.SetDealSupplierID(dealID, ordersCrud.GetOrderAuthor(_askID));
        dealsCrud.SetDealMasterID(dealID, administratum.GetMaster(ordersCrud.GetOrderAuthor(_askID)));
        dealsCrud.SetDealAskID(dealID, _askID);
        dealsCrud.SetDealBidID(dealID, _bidID);
        dealsCrud.SetDealDuration(dealID, ordersCrud.GetOrderDuration(_bidID));
        dealsCrud.SetDealPrice(dealID, ordersCrud.GetOrderPrice(_askID));
        dealsCrud.SetDealStartTime(dealID, block.timestamp);
        dealsCrud.SetDealEndTime(dealID, endTime);
        dealsCrud.SetDealStatus(dealID, Deals.DealStatus.STATUS_ACCEPTED);
        dealsCrud.SetDealBlockedBalance(dealID, ordersCrud.GetOrderFrozenSum(_bidID));
        dealsCrud.SetDealTotalPayout(dealID, 0);
        dealsCrud.SetDealLastBillTS(dealID, block.timestamp);



        emit DealOpened(dealsCrud.GetDealsAmount());

        ordersCrud.SetOrderDealID(_askID, dealID);
        ordersCrud.SetOrderDealID(_bidID, dealID);
    }

    function CloseDeal(uint dealID, BlacklistPerson blacklisted) public returns (bool){
        require((dealsCrud.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED));
        require(msg.sender == dealsCrud.GetDealSupplierID(dealID) || msg.sender == dealsCrud.GetDealConsumerID(dealID) || msg.sender == dealsCrud.GetDealMasterID(dealID));

        if (block.timestamp <= dealsCrud.GetDealStartTime(dealID).add(dealsCrud.GetDealDuration(dealID))) {
            // after endTime
            require(dealsCrud.GetDealConsumerID(dealID) == msg.sender);
        }
        AddToBlacklist(dealID, blacklisted);
        InternalBill(dealID);
        InternalCloseDeal(dealID);
        RefundRemainingFunds(dealID);
        return true;
    }

    function Bill(uint dealID) public returns (bool){
        InternalBill(dealID);
        ReserveNextPeriodFunds(dealID);
        return true;
    }

    function CreateChangeRequest(uint dealID, uint newPrice, uint newDuration) public returns (uint changeRequestID) {

        require(
            msg.sender == dealsCrud.GetDealConsumerID(dealID)
            || msg.sender ==  dealsCrud.GetDealMasterID(dealID)
            || msg.sender == dealsCrud.GetDealSupplierID(dealID));
        require(dealsCrud.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED);

        if (dealsCrud.GetDealDuration(dealID) == 0) {
            require(newDuration == 0);
        }


        Orders.OrderType requestType;

        if (msg.sender == dealsCrud.GetDealConsumerID(dealID)) {
            requestType = Orders.OrderType.ORDER_BID;
        } else {
            requestType = Orders.OrderType.ORDER_ASK;
        }

        uint currentID = changeRequestsCrud.Write(dealID, requestType, newPrice, newDuration, ChangeRequests.RequestStatus.REQUEST_CREATED);
        emit DealChangeRequestSet(currentID);

        if (requestType == Orders.OrderType.ORDER_BID) {
            uint oldID = changeRequestsCrud.GetActualChangeRequest(dealID, 1);
            uint matchingID = changeRequestsCrud.GetActualChangeRequest(dealID, 0);
            emit DealChangeRequestUpdated(oldID);
            changeRequestsCrud.SetChangeRequestStatus(oldID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            changeRequestsCrud.SetActualChangeRequest(dealID, 1, currentID);

            if (newDuration == dealsCrud.GetDealDuration(dealID) && newPrice > dealsCrud.GetDealPrice(dealID)) {
                if (changeRequestsCrud.GetChangeRequestPrice(matchingID) <= newPrice) {
                    changeRequestsCrud.SetChangeRequestStatus(matchingID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                }
                changeRequestsCrud.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                Bill(dealID);
                dealsCrud.SetDealPrice(dealID, newPrice);
                changeRequestsCrud.SetActualChangeRequest(dealID, 1, 0);
                emit DealChangeRequestUpdated(requestsAmount);
            } else if (changeRequestsCrud.GetChangeRequestStatus(matchingID) == ChangeRequests.RequestStatus.REQUEST_CREATED
                    && changeRequestsCrud.GetChangeRequestDuration(matchingID) >= newDuration
                    && changeRequestsCrud.GetChangeRequestPrice(matchingID) <= newPrice) {

                changeRequestsCrud.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                changeRequestsCrud.SetChangeRequestStatus(matchingID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                emit DealChangeRequestUpdated(changeRequestsCrud.GetActualChangeRequest(dealID, 0));
                changeRequestsCrud.SetActualChangeRequest(dealID, 0, 0);
                changeRequestsCrud.SetActualChangeRequest(dealID, 1, 0);
                Bill(dealID);
                dealsCrud.SetDealPrice(dealID, changeRequestsCrud.GetChangeRequestPrice(matchingID));
                dealsCrud.SetDealDuration(dealID, newDuration);
                emit DealChangeRequestUpdated(requestsAmount);
            } else {
                return currentID;
            }
            changeRequestsCrud.SetChangeRequestStatus(changeRequestsCrud.GetActualChangeRequest(dealID, 1), ChangeRequests.RequestStatus.REQUEST_CANCELED);
            emit DealChangeRequestUpdated(changeRequestsCrud.GetActualChangeRequest(dealID, 1));
            changeRequestsCrud.SetActualChangeRequest(dealID, 1, changeRequestsCrud.GetChangeRequestsAmount());
        }

        if (requestType == Orders.OrderType.ORDER_ASK) {
            matchingID = changeRequestsCrud.GetActualChangeRequest(dealID, 1);
            oldID = changeRequestsCrud.GetActualChangeRequest(dealID, 0);
            emit DealChangeRequestUpdated(oldID);
            changeRequestsCrud.SetChangeRequestStatus(oldID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            changeRequestsCrud.SetActualChangeRequest(dealID, 0, currentID);


            if (newDuration == dealsCrud.GetDealDuration(dealID) && newPrice < dealsCrud.GetDealPrice(dealID)) {
                if (changeRequestsCrud.GetChangeRequestPrice(matchingID) >= newPrice) {
                    changeRequestsCrud.SetChangeRequestStatus(matchingID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                }
                changeRequestsCrud.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                Bill(dealID);
                dealsCrud.SetDealPrice(dealID, newPrice);
                changeRequestsCrud.SetActualChangeRequest(dealID, 0, 0);
                emit DealChangeRequestUpdated(currentID);
            } else if (changeRequestsCrud.GetChangeRequestStatus(matchingID) == ChangeRequests.RequestStatus.REQUEST_CREATED
                && changeRequestsCrud.GetChangeRequestDuration(matchingID) <= newDuration
                && changeRequestsCrud.GetChangeRequestPrice(matchingID) >= newPrice) {

                changeRequestsCrud.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                changeRequestsCrud.SetChangeRequestStatus(matchingID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                emit DealChangeRequestUpdated(matchingID); //bug
                changeRequestsCrud.SetActualChangeRequest(dealID, 0, 0);
                changeRequestsCrud.SetActualChangeRequest(dealID, 1, 0);
                Bill(dealID);
                dealsCrud.SetDealPrice(dealID, newPrice);
                dealsCrud.SetDealDuration(dealID, changeRequestsCrud.GetChangeRequestDuration(matchingID));
                emit DealChangeRequestUpdated(currentID);
            } else {
                return currentID;
            }
        }

        dealsCrud.SetDealEndTime(dealID, dealsCrud.GetDealStartTime(dealID).add(dealsCrud.GetDealEndTime(dealID)));
        return currentID;
    }

    function CancelChangeRequest(uint changeRequestID) public returns (bool) {
        uint dealID = changeRequestsCrud.GetChangeRequestDealID(changeRequestID);
        require(msg.sender == dealsCrud.GetDealSupplierID(dealID) || msg.sender == dealsCrud.GetDealMasterID(dealID) || msg.sender == dealsCrud.GetDealConsumerID(dealID));
        require(changeRequestsCrud.GetChangeRequestStatus(changeRequestID) != ChangeRequests.RequestStatus.REQUEST_ACCEPTED);

        if (changeRequestsCrud.GetChangeRequestType(changeRequestID) == Orders.OrderType.ORDER_ASK) {
            if (msg.sender == dealsCrud.GetDealConsumerID(dealID)) {
                changeRequestsCrud.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_REJECTED);
            } else {
                changeRequestsCrud.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            }
            changeRequestsCrud.SetActualChangeRequest(dealID, 0, 0);
            emit DealChangeRequestUpdated(changeRequestID);
        }

        if (changeRequestsCrud.GetChangeRequestType(changeRequestID) == Orders.OrderType.ORDER_BID) {
            if (msg.sender == dealsCrud.GetDealConsumerID(dealID)) {
                changeRequestsCrud.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            } else {
                changeRequestsCrud.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_REJECTED);
            }
            changeRequestsCrud.SetActualChangeRequest(dealID, 1, 0);
            emit DealChangeRequestUpdated(changeRequestID);
        }
        return true;
    }

    // GETTERS

    function GetBenchmarksQuantity() public view returns (uint) {
        return benchmarksQuantity;
    }

    function GetNetflagsQuantity() public view returns (uint) {
        return netflagsQuantity;
    }


    // INTERNAL

    function InternalBill(uint dealID) internal returns (bool){

        require(dealsCrud.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED);
        require(
            msg.sender == dealsCrud.GetDealSupplierID(dealID) ||
            msg.sender == dealsCrud.GetDealConsumerID(dealID) ||
            msg.sender == dealsCrud.GetDealMasterID(dealID));
        uint paidAmount;

        if (dealsCrud.GetDealDuration(dealID) != 0 && dealsCrud.GetDealLastBillTS(dealID) >= dealsCrud.GetDealEndTime(dealID)) {
            // means we already billed deal after endTime
            return true;
        } else if (dealsCrud.GetDealDuration(dealID) != 0
            && block.timestamp > dealsCrud.GetDealEndTime(dealID)
            && dealsCrud.GetDealLastBillTS(dealID) < dealsCrud.GetDealEndTime(dealID)) {
            paidAmount = CalculatePayment(dealsCrud.GetDealPrice(dealID), dealsCrud.GetDealEndTime(dealID).sub(dealsCrud.GetDealLastBillTS(dealID)));
        } else {
            paidAmount = CalculatePayment(dealsCrud.GetDealPrice(dealID), block.timestamp.sub(dealsCrud.GetDealLastBillTS(dealID)));
        }

        if (paidAmount > dealsCrud.GetDealBlockedBalance(dealID)) {
            if (token.balanceOf(dealsCrud.GetDealConsumerID(dealID)) >= paidAmount.sub(dealsCrud.GetDealBlockedBalance(dealID))) {
                require(token.transferFrom(dealsCrud.GetDealConsumerID(dealID), this, paidAmount.sub(dealsCrud.GetDealBlockedBalance(dealID))));
                dealsCrud.SetDealBlockedBalance(dealID, dealsCrud.GetDealBlockedBalance(dealID).add(paidAmount.sub(dealsCrud.GetDealBlockedBalance(dealID))));
            } else {
                emit Billed(dealID, dealsCrud.GetDealBlockedBalance(dealID));
                InternalCloseDeal(dealID);
                require(token.transfer(dealsCrud.GetDealMasterID(dealID), dealsCrud.GetDealBlockedBalance(dealID)));
                dealsCrud.SetDealTotalPayout(dealID, dealsCrud.GetDealTotalPayout(dealID).add(dealsCrud.GetDealBlockedBalance(dealID)));
                dealsCrud.SetDealBlockedBalance(dealID, 0);
                dealsCrud.SetDealEndTime(dealID, block.timestamp);
                return true;
            }
        }

        require(token.transfer(dealsCrud.GetDealMasterID(dealID), paidAmount));
        dealsCrud.SetDealBlockedBalance(dealID, dealsCrud.GetDealBlockedBalance(dealID).sub(paidAmount));
        dealsCrud.SetDealTotalPayout(dealID, dealsCrud.GetDealTotalPayout(dealID).add(paidAmount));
        dealsCrud.SetDealLastBillTS(dealID, block.timestamp);
        emit Billed(dealID, paidAmount);
        return true;
    }

    function ReserveNextPeriodFunds(uint dealID) internal returns (bool) {
        uint nextPeriod;
        address consumerID = dealsCrud.GetDealConsumerID(dealID);

        (uint duration, uint price, uint endTime, Deals.DealStatus status, uint blockedBalance, , ) = dealsCrud.GetDealParams(dealID);

        if (duration == 0) {
            if (status == Deals.DealStatus.STATUS_CLOSED) {
                return true;
            }
            nextPeriod = 1 hours;
        } else {
            if (block.timestamp > endTime) {
                // we don't reserve funds for next period
                return true;
            }
            if (endTime.sub(block.timestamp) < 1 days) {
                nextPeriod = endTime.sub(block.timestamp);
            } else {
                nextPeriod = 1 days;
            }
        }

        if (CalculatePayment(price, nextPeriod) > blockedBalance) {
            uint nextPeriodSum = CalculatePayment(price, nextPeriod).sub(blockedBalance);

            if (token.balanceOf(consumerID) >= nextPeriodSum) {
                require(token.transferFrom(consumerID, this, nextPeriodSum));
                dealsCrud.SetDealBlockedBalance(dealID, blockedBalance.add(nextPeriodSum));
            } else {
                emit Billed(dealID, blockedBalance);
                InternalCloseDeal(dealID);
                RefundRemainingFunds(dealID);
                return true;
            }
        }
        return true;
    }

    function RefundRemainingFunds(uint dealID) internal returns (bool) {
        address consumerID = dealsCrud.GetDealConsumerID(dealID);

        uint blockedBalance = dealsCrud.GetDealBlockedBalance(dealID);

        if (blockedBalance != 0) {
            token.transfer(consumerID, blockedBalance);
            dealsCrud.SetDealBlockedBalance(dealID, 0);
        }
        return true;
    }



    function CalculatePayment(uint _price, uint _period) internal view returns (uint) {
        uint rate = oracle.getCurrentPrice();
        return rate.mul(_price).mul(_period).div(1e18);
    }

    function AddToBlacklist(uint dealID, BlacklistPerson role) internal {
        (, address supplierID, address consumerID, address masterID, , , ) = dealsCrud.GetDealInfo(dealID);

        // only consumer can blacklist
        require(msg.sender == consumerID || role == BlacklistPerson.BLACKLIST_NOBODY);
        if (role == BlacklistPerson.BLACKLIST_WORKER) {
            bl.Add(consumerID, supplierID);
        } else if (role == BlacklistPerson.BLACKLIST_MASTER) {
            bl.Add(consumerID, masterID);
        }
    }

    function InternalCloseDeal(uint dealID) internal {
        ( , address supplierID, address consumerID, address masterID, , , ) = dealsCrud.GetDealInfo(dealID);

        Deals.DealStatus status = dealsCrud.GetDealStatus(dealID);

        if (status == Deals.DealStatus.STATUS_CLOSED) {
            return;
        }
        require((status == Deals.DealStatus.STATUS_ACCEPTED));
        require(msg.sender == consumerID || msg.sender == supplierID || msg.sender == masterID);
        dealsCrud.SetDealStatus(dealID, Deals.DealStatus.STATUS_CLOSED);
        dealsCrud.SetDealEndTime(dealID, block.timestamp);
        emit DealUpdated(dealID);
    }

    function ResizeBenchmarks(uint64[] _benchmarks) internal view returns (uint64[]) {
        uint64[] memory benchmarks = new uint64[](benchmarksQuantity);
        for (uint i = 0; i < _benchmarks.length; i++) {
            benchmarks[i] = _benchmarks[i];
        }
        return benchmarks;
    }

    function ResizeNetflags(bool[] _netflags) internal view returns (bool[]) {
        bool[] memory netflags = new bool[](netflagsQuantity);
        for (uint i = 0; i < _netflags.length; i++) {
            netflags[i] = _netflags[i];
        }
        return netflags;
    }

    // SETTERS

    function SetProfileRegistryAddress(address _newPR) public onlyOwner returns (bool) {
        profiles = ProfileRegistry(_newPR);
        return true;
    }

    function SetBlacklistAddress(address _newBL) public onlyOwner returns (bool) {
        bl = Blacklist(_newBL);
        return true;
    }

    function SetOracleAddress(address _newOracle) public onlyOwner returns (bool) {
        require(OracleUSD(_newOracle).getCurrentPrice() != 0);
        oracle = OracleUSD(_newOracle);
        return true;
    }

    function SetDealsAddress(address _newDeals) public onlyOwner returns (bool) {
        dealsCrud = Deals(_newDeals);
        return true;
    }

    function SetOrdersAddress(address _newOrders) public onlyOwner returns (bool) {
        ordersCrud = Orders(_newOrders);
        return  true;
    }

    function SetBenchmarksQuantity(uint _newQuantity) public onlyOwner returns (bool) {
        require(_newQuantity > benchmarksQuantity);
        emit NumBenchmarksUpdated(_newQuantity);
        benchmarksQuantity = _newQuantity;
        return true;
    }

    function SetNetflagsQuantity(uint _newQuantity) public onlyOwner returns (bool) {
        require(_newQuantity > netflagsQuantity);
        emit NumNetflagsUpdated(_newQuantity);
        netflagsQuantity = _newQuantity;
        return true;
    }

    function KillMarket() onlyOwner public {
        token.transfer(owner, token.balanceOf(address(this)));
        selfdestruct(owner);
    }

    function Migrate(address _newMarket) public onlyOwner returns (bool) {
        token.transfer(_newMarket, token.balanceOf(address(this)));
        ordersCrud.transferOwnership(_newMarket);
        dealsCrud.transferOwnership(_newMarket);
        changeRequestsCrud.transferOwnership(_newMarket);
        pause();

    }

    // OLD API

    function RegisterWorker(address _master) public whenNotPaused returns (bool) {
        administratum.ExternalRegisterWorker(_master, msg.sender);
        emit WorkerAnnounced(msg.sender, _master);
        return true;
    }

    function ConfirmWorker(address _worker) public whenNotPaused returns (bool) {
        administratum.ExternalConfirmWorker(_worker, msg.sender);
        emit WorkerConfirmed(_worker, msg.sender);
        return true;
    }

    function RemoveWorker(address _worker, address _master) public whenNotPaused returns (bool) {
        administratum.ExternalRemoveWorker(_worker, _master, msg.sender);
        emit WorkerRemoved(_worker, _master);
        return true;
    }

    function GetDealsAmount() public view returns (uint){
        return dealsCrud.GetDealsAmount();
    }

    function GetOrdersAmount() public view returns (uint){
        return ordersCrud.GetOrdersAmount();
    }

    function GetChangeRequestsAmount() public view returns (uint){
        return changeRequestsCrud.GetChangeRequestsAmount();
    }

    function GetMaster(address _worker) view public returns (address master) {
        return administratum.GetMaster(_worker);
    }


    function GetChangeRequestInfo(uint changeRequestID) view public
    returns (
        uint dealID,
        Orders.OrderType requestType,
        uint price,
        uint duration,
        ChangeRequests.RequestStatus status
    ){
        return (
        changeRequestsCrud.GetChangeRequestDealID(changeRequestID),
        changeRequestsCrud.GetChangeRequestType(changeRequestID),
        changeRequestsCrud.GetChangeRequestPrice(changeRequestID),
        changeRequestsCrud.GetChangeRequestDuration(changeRequestID),
        changeRequestsCrud.GetChangeRequestStatus(changeRequestID)
        );
    }

    function GetOrderInfo(uint orderID) view public
    returns (
        Orders.OrderType orderType,
        address author,
        address counterparty,
        uint duration,
        uint price,
        bool[] netflags,
        ProfileRegistry.IdentityLevel identityLevel,
        address blacklist,
        bytes32 tag,
        uint64[] benchmarks,
        uint frozenSum
    ){
        orderType = ordersCrud.GetOrderType(orderID);
        author = ordersCrud.GetOrderAuthor(orderID);
        counterparty = ordersCrud.GetOrderCounterparty(orderID);
        duration =  ordersCrud.GetOrderDuration(orderID);
        price = ordersCrud.GetOrderPrice(orderID);
        netflags = ordersCrud.GetOrderNetflags(orderID);
        identityLevel = ordersCrud.GetOrderIdentityLevel(orderID);
        blacklist = ordersCrud.GetOrderBlacklist(orderID);
        tag = ordersCrud.GetOrderTag(orderID);
        benchmarks = ordersCrud.GetOrderBenchmarks(orderID);
        frozenSum = ordersCrud.GetOrderFrozenSum(orderID);
    }


    function GetOrderParams(uint orderID) view public
    returns (
        Orders.OrderStatus orderStatus,
        uint dealID
    ){
        orderStatus = ordersCrud.GetOrderStatus(orderID);
        dealID = ordersCrud.GetOrderDealID(orderID);
    }

    function GetDealInfo(uint dealID) view public
    returns (
        uint64[] benchmarks,
        address supplierID,
        address consumerID,
        address masterID,
        uint askID,
        uint bidID,
        uint startTime
    ){
        benchmarks = dealsCrud.GetDealBenchmarks(dealID);
        supplierID = dealsCrud.GetDealSupplierID(dealID);
        consumerID = dealsCrud.GetDealConsumerID(dealID);
        masterID = dealsCrud.GetDealMasterID(dealID);
        askID = dealsCrud.GetDealAskID(dealID);
        bidID = dealsCrud.GetDealBidID(dealID);
        startTime = dealsCrud.GetDealStartTime(dealID);
    }

    function GetDealParams(uint dealID) view public
    returns (
        uint duration,
        uint price,
        uint endTime,
        Deals.DealStatus status,
        uint blockedBalance,
        uint totalPayout,
        uint lastBillTS
    ){
        duration = dealsCrud.GetDealDuration(dealID);
        price = dealsCrud.GetDealPrice(dealID);
        endTime = dealsCrud.GetDealEndTime(dealID);
        status = dealsCrud.GetDealStatus(dealID);
        blockedBalance = dealsCrud.GetDealBlockedBalance(dealID);
        totalPayout = dealsCrud.GetDealTotalPayout(dealID);
        lastBillTS = dealsCrud.GetDealLastBillTS(dealID);
    }

}
