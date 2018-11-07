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

    ProfileRegistry pr;

    Orders ord;

    Administratum adm;

    Deals dl;

    ChangeRequests cr;

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
        uint _netflagsQuantity) public {
        token = SNM(_token);
        bl = Blacklist(_blacklist);
        oracle = OracleUSD(_oracle);
        pr = ProfileRegistry(_profileRegistry);
        adm = Administratum(_administratum);
        ord = Orders(_orders);
        dl = Deals(_deals);
        cr = ChangeRequests(_changeRequests);
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


        ord.Write(
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

        emit OrderPlaced(ord.GetOrdersAmount());

        return ord.GetOrdersAmount();
    }

    function CancelOrder(uint orderID) public returns (bool) {
        require(orderID <= ord.GetOrdersAmount());

        address author = ord.GetOrderAuthor(orderID);

        uint frozenSum = ord.GetOrderFrozenSum(orderID);

        Orders.OrderStatus orderStatus = ord.GetOrderStatus(orderID);


        require(orderStatus == Orders.OrderStatus.ORDER_ACTIVE);
        require(author == msg.sender || adm.GetMaster(author) == msg.sender);

        require(token.transfer(msg.sender, frozenSum));

        ord.SetOrderStatus(orderID, Orders.OrderStatus.ORDER_INACTIVE);

        emit OrderUpdated(orderID);

        return true;
    }

    function QuickBuy(uint askID, uint buyoutDuration) public whenNotPaused {

        require(ord.GetOrderType(askID) == Orders.OrderType.ORDER_ASK);
        require(ord.GetOrderStatus(askID) == Orders.OrderStatus.ORDER_ACTIVE);

        require(ord.GetOrderDuration(askID) >= buyoutDuration);
        require(pr.GetProfileLevel(msg.sender) >= ord.GetOrderIdentityLevel(askID));
        require(bl.Check(ord.GetOrderBlacklist(askID), msg.sender) == false);
        require(
            bl.Check(msg.sender, adm.GetMaster(ord.GetOrderAuthor(askID))) == false
            && bl.Check(ord.GetOrderAuthor(askID), msg.sender) == false);

        PlaceOrder(
            Orders.OrderType.ORDER_BID,
            adm.GetMaster(ord.GetOrderAuthor(askID)),
            buyoutDuration,
            ord.GetOrderPrice(askID),
            ord.GetOrderNetflags(askID),
            ProfileRegistry.IdentityLevel.ANONYMOUS,
            address(0),
            bytes32(0),
            ord.GetOrderBenchmarks(askID));

        OpenDeal(askID, ord.GetOrdersAmount());
    }

    // Deal functions

    function OpenDeal(uint _askID, uint _bidID) whenNotPaused public {
        require(
            ord.GetOrderStatus(_askID) == Orders.OrderStatus.ORDER_ACTIVE
            && ord.GetOrderStatus(_bidID) == Orders.OrderStatus.ORDER_ACTIVE);
        require(ord.GetOrderCounterparty(_askID) == 0x0 || ord.GetOrderCounterparty(_askID) == adm.GetMaster(ord.GetOrderAuthor(_bidID)));
        require(ord.GetOrderCounterparty(_bidID) == 0x0 || ord.GetOrderCounterparty(_bidID) == adm.GetMaster(ord.GetOrderAuthor(_askID)));
        require(ord.GetOrderType(_askID) == Orders.OrderType.ORDER_ASK);
        require(ord.GetOrderType(_bidID) == Orders.OrderType.ORDER_BID);
        require(!bl.Check(ord.GetOrderBlacklist(_bidID), adm.GetMaster(ord.GetOrderAuthor(_askID))));
        require(!bl.Check(ord.GetOrderBlacklist(_bidID), ord.GetOrderAuthor(_askID)));
        require(!bl.Check(ord.GetOrderAuthor(_bidID), adm.GetMaster(ord.GetOrderAuthor(_askID))));
        require(!bl.Check(ord.GetOrderAuthor(_bidID), ord.GetOrderAuthor(_askID)));
        require(!bl.Check(ord.GetOrderBlacklist(_askID), ord.GetOrderAuthor(_bidID)));
        require(!bl.Check(adm.GetMaster(ord.GetOrderAuthor(_askID)), ord.GetOrderAuthor(_bidID)));
        require(!bl.Check(ord.GetOrderAuthor(_askID), ord.GetOrderAuthor(_bidID)));
        require(ord.GetOrderPrice(_askID) <= ord.GetOrderPrice(_bidID));
        require(ord.GetOrderDuration(_askID) >= ord.GetOrderDuration(_bidID));
        // profile level check
        require(pr.GetProfileLevel(ord.GetOrderAuthor(_bidID)) >= ord.GetOrderIdentityLevel(_askID));
        require(pr.GetProfileLevel(adm.GetMaster(ord.GetOrderAuthor(_askID))) >= ord.GetOrderIdentityLevel(_bidID)); //bug

        bool[] memory askNetflags = ord.GetOrderNetflags(_askID);
        if (askNetflags.length < netflagsQuantity) {
            askNetflags = ResizeNetflags(askNetflags);
        }

        bool[] memory bidNetflags = ord.GetOrderNetflags(_bidID);
        if (bidNetflags.length < netflagsQuantity) {
            bidNetflags = ResizeNetflags(bidNetflags);
        }

        for (uint i = 0; i < netflagsQuantity; i++) {
            // implementation: when bid contains requirement, ask necessary needs to have this
            // if ask have this one - pass
            require(!bidNetflags[i] || askNetflags[i]);
        }

        uint64[] memory askBenchmarks = ord.GetOrderBenchmarks(_askID);
        if (askBenchmarks.length < benchmarksQuantity) {
            askBenchmarks = ResizeBenchmarks(askBenchmarks);
        }

        uint64[] memory bidBenchmarks = ord.GetOrderBenchmarks(_bidID);
        if (bidBenchmarks.length < benchmarksQuantity) {
            bidBenchmarks = ResizeBenchmarks(bidBenchmarks);
        }

        for (i = 0; i < benchmarksQuantity; i++) {
            require(askBenchmarks[i] >= bidBenchmarks[i]);
        }

        ord.SetOrderStatus(_askID, Orders.OrderStatus.ORDER_INACTIVE);
        ord.SetOrderStatus(_bidID, Orders.OrderStatus.ORDER_INACTIVE);

        emit OrderUpdated(_askID);
        emit OrderUpdated(_bidID);


        // if deal is normal
        if (ord.GetOrderDuration(_askID) != 0) {
            uint endTime = block.timestamp.add(ord.GetOrderDuration(_bidID));
        } else {
            endTime = 0;
            // `0` - for spot deal
        }

        dl.IncreaseDealsAmount();

        uint dealID = dl.GetDealsAmount();

        dl.SetDealBenchmarks(dealID, askBenchmarks);
        dl.SetDealConsumerID(dealID, ord.GetOrderAuthor(_bidID));
        dl.SetDealSupplierID(dealID, ord.GetOrderAuthor(_askID));
        dl.SetDealMasterID(dealID, adm.GetMaster(ord.GetOrderAuthor(_askID)));
        dl.SetDealAskID(dealID, _askID);
        dl.SetDealBidID(dealID, _bidID);
        dl.SetDealDuration(dealID, ord.GetOrderDuration(_bidID));
        dl.SetDealPrice(dealID, ord.GetOrderPrice(_askID));
        dl.SetDealStartTime(dealID, block.timestamp);
        dl.SetDealEndTime(dealID, endTime);
        dl.SetDealStatus(dealID, Deals.DealStatus.STATUS_ACCEPTED);
        dl.SetDealBlockedBalance(dealID, ord.GetOrderFrozenSum(_bidID));
        dl.SetDealTotalPayout(dealID, 0);
        dl.SetDealLastBillTS(dealID, block.timestamp);



        emit DealOpened(dl.GetDealsAmount());

        ord.SetOrderDealID(_askID, dealID);
        ord.SetOrderDealID(_bidID, dealID);
    }

    function CloseDeal(uint dealID, BlacklistPerson blacklisted) public returns (bool){
        require((dl.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED));
        require(msg.sender == dl.GetDealSupplierID(dealID) || msg.sender == dl.GetDealConsumerID(dealID) || msg.sender == dl.GetDealMasterID(dealID));

        if (block.timestamp <= dl.GetDealStartTime(dealID).add(dl.GetDealDuration(dealID))) {
            // after endTime
            require(dl.GetDealConsumerID(dealID) == msg.sender);
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
            msg.sender == dl.GetDealConsumerID(dealID)
            || msg.sender ==  dl.GetDealMasterID(dealID)
            || msg.sender == dl.GetDealSupplierID(dealID));
        require(dl.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED);

        if (dl.GetDealDuration(dealID) == 0) {
            require(newDuration == 0);
        }


        Orders.OrderType requestType;

        if (msg.sender == dl.GetDealConsumerID(dealID)) {
            requestType = Orders.OrderType.ORDER_BID;
        } else {
            requestType = Orders.OrderType.ORDER_ASK;
        }

        uint currentID = cr.Write(dealID, requestType, newPrice, newDuration, ChangeRequests.RequestStatus.REQUEST_CREATED);
        emit DealChangeRequestSet(currentID);

        if (requestType == Orders.OrderType.ORDER_BID) {
            uint oldID = cr.GetActualChangeRequest(dealID, 1);
            uint matchingID = cr.GetActualChangeRequest(dealID, 0);
            emit DealChangeRequestUpdated(oldID);
            cr.SetChangeRequestStatus(oldID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            cr.SetActualChangeRequest(dealID, 1, currentID);

            if (newDuration == dl.GetDealDuration(dealID) && newPrice > dl.GetDealPrice(dealID)) {
                cr.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                Bill(dealID);
                dl.SetDealPrice(dealID, newPrice);
                cr.SetActualChangeRequest(dealID, 1, 0);
                emit DealChangeRequestUpdated(requestsAmount);
            } else if (cr.GetChangeRequestStatus(matchingID) == ChangeRequests.RequestStatus.REQUEST_CREATED
                    && cr.GetChangeRequestDuration(matchingID) >= newDuration
                    && cr.GetChangeRequestPrice(matchingID) <= newPrice) {

                cr.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                cr.SetChangeRequestStatus(matchingID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                emit DealChangeRequestUpdated(cr.GetActualChangeRequest(dealID, 0));
                cr.SetActualChangeRequest(dealID, 0, 0);
                cr.SetActualChangeRequest(dealID, 1, 0);
                Bill(dealID);
                dl.SetDealPrice(dealID, cr.GetChangeRequestPrice(matchingID));
                dl.SetDealDuration(dealID, newDuration);
                emit DealChangeRequestUpdated(requestsAmount);
            } else {
                return currentID;
            }
            cr.SetChangeRequestStatus(cr.GetActualChangeRequest(dealID, 1), ChangeRequests.RequestStatus.REQUEST_CANCELED);
            emit DealChangeRequestUpdated(cr.GetActualChangeRequest(dealID, 1));
            cr.SetActualChangeRequest(dealID, 1, cr.GetChangeRequestsAmount());
        }

        if (requestType == Orders.OrderType.ORDER_ASK) {
            matchingID = cr.GetActualChangeRequest(dealID, 1);
            oldID = cr.GetActualChangeRequest(dealID, 0);
            emit DealChangeRequestUpdated(oldID);
            cr.SetChangeRequestStatus(oldID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            cr.SetActualChangeRequest(dealID, 0, currentID);


            if (newDuration == dl.GetDealDuration(dealID) && newPrice < dl.GetDealPrice(dealID)) {
                cr.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                Bill(dealID);
                dl.SetDealPrice(dealID, newPrice);
                cr.SetActualChangeRequest(dealID, 0, 0);
                emit DealChangeRequestUpdated(currentID);
            } else if (cr.GetChangeRequestStatus(matchingID) == ChangeRequests.RequestStatus.REQUEST_CREATED
                && cr.GetChangeRequestDuration(matchingID) <= newDuration
                && cr.GetChangeRequestPrice(matchingID) >= newPrice) {

                cr.SetChangeRequestStatus(currentID, ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                cr.SetChangeRequestStatus(cr.GetActualChangeRequest(dealID, 1), ChangeRequests.RequestStatus.REQUEST_ACCEPTED);
                emit DealChangeRequestUpdated(matchingID); //bug
                cr.SetActualChangeRequest(dealID, 0, 0);
                cr.SetActualChangeRequest(dealID, 1, 0);
                Bill(dealID);
                dl.SetDealPrice(dealID, newPrice);
                dl.SetDealDuration(dealID, cr.GetChangeRequestDuration(matchingID));
                emit DealChangeRequestUpdated(currentID);
            } else {
                return currentID;
            }
        }

        dl.SetDealEndTime(dealID, dl.GetDealStartTime(dealID).add(dl.GetDealEndTime(dealID)));
        return currentID;
    }

    function CancelChangeRequest(uint changeRequestID) public returns (bool) {
        uint dealID = cr.GetChangeRequestDealID(changeRequestID);
        require(msg.sender == dl.GetDealSupplierID(dealID) || msg.sender == dl.GetDealMasterID(dealID) || msg.sender == dl.GetDealConsumerID(dealID));
        require(cr.GetChangeRequestStatus(changeRequestID) != ChangeRequests.RequestStatus.REQUEST_ACCEPTED);

        if (cr.GetChangeRequestType(changeRequestID) == Orders.OrderType.ORDER_ASK) {
            if (msg.sender == dl.GetDealConsumerID(dealID)) {
                cr.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_REJECTED);
            } else {
                cr.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            }
            cr.SetActualChangeRequest(dealID, 0, 0);
            emit DealChangeRequestUpdated(changeRequestID);
        }

        if (cr.GetChangeRequestType(changeRequestID) == Orders.OrderType.ORDER_BID) {
            if (msg.sender == dl.GetDealConsumerID(dealID)) {
                cr.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_CANCELED);
            } else {
                cr.SetChangeRequestStatus(changeRequestID, ChangeRequests.RequestStatus.REQUEST_REJECTED);
            }
            cr.SetActualChangeRequest(dealID, 1, 0);
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

        require(dl.GetDealStatus(dealID) == Deals.DealStatus.STATUS_ACCEPTED);
        require(
            msg.sender == dl.GetDealSupplierID(dealID) ||
            msg.sender == dl.GetDealConsumerID(dealID) ||
            msg.sender == dl.GetDealMasterID(dealID));
        uint paidAmount;

        if (dl.GetDealDuration(dealID) != 0 && dl.GetDealLastBillTS(dealID) >= dl.GetDealEndTime(dealID)) {
            // means we already billed deal after endTime
            return true;
        } else if (dl.GetDealDuration(dealID) != 0
            && block.timestamp > dl.GetDealEndTime(dealID)
            && dl.GetDealLastBillTS(dealID) < dl.GetDealEndTime(dealID)) {
            paidAmount = CalculatePayment(dl.GetDealPrice(dealID), dl.GetDealEndTime(dealID).sub(dl.GetDealLastBillTS(dealID)));
        } else {
            paidAmount = CalculatePayment(dl.GetDealPrice(dealID), block.timestamp.sub(dl.GetDealLastBillTS(dealID)));
        }

        if (paidAmount > dl.GetDealBlockedBalance(dealID)) {
            if (token.balanceOf(dl.GetDealConsumerID(dealID)) >= paidAmount.sub(dl.GetDealBlockedBalance(dealID))) {
                require(token.transferFrom(dl.GetDealConsumerID(dealID), this, paidAmount.sub(dl.GetDealBlockedBalance(dealID))));
                dl.SetDealBlockedBalance(dealID, dl.GetDealBlockedBalance(dealID).add(paidAmount.sub(dl.GetDealBlockedBalance(dealID))));
            } else {
                emit Billed(dealID, dl.GetDealBlockedBalance(dealID));
                InternalCloseDeal(dealID);
                require(token.transfer(dl.GetDealMasterID(dealID), dl.GetDealBlockedBalance(dealID)));
                dl.SetDealTotalPayout(dealID, dl.GetDealTotalPayout(dealID).add(dl.GetDealBlockedBalance(dealID)));
                dl.SetDealBlockedBalance(dealID, 0);
                dl.SetDealEndTime(dealID, block.timestamp);
                return true;
            }
        }

        require(token.transfer(dl.GetDealMasterID(dealID), paidAmount));
        dl.SetDealBlockedBalance(dealID, dl.GetDealBlockedBalance(dealID).sub(paidAmount));
        dl.SetDealTotalPayout(dealID, dl.GetDealTotalPayout(dealID).add(paidAmount));
        dl.SetDealLastBillTS(dealID, block.timestamp);
        emit Billed(dealID, paidAmount);
        return true;
    }

    function ReserveNextPeriodFunds(uint dealID) internal returns (bool) {
        uint nextPeriod;
        address consumerID = dl.GetDealConsumerID(dealID);

        (uint duration, uint price, uint endTime, Deals.DealStatus status, uint blockedBalance, , ) = dl.GetDealParams(dealID);

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
                dl.SetDealBlockedBalance(dealID, blockedBalance.add(nextPeriodSum));
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
        address consumerID = dl.GetDealConsumerID(dealID);

        uint blockedBalance = dl.GetDealBlockedBalance(dealID);

        if (blockedBalance != 0) {
            token.transfer(consumerID, blockedBalance);
            dl.SetDealBlockedBalance(dealID, 0);
        }
        return true;
    }



    function CalculatePayment(uint _price, uint _period) internal view returns (uint) {
        uint rate = oracle.getCurrentPrice();
        return rate.mul(_price).mul(_period).div(1e18);
    }

    function AddToBlacklist(uint dealID, BlacklistPerson role) internal {
        (, address supplierID, address consumerID, address masterID, , , ) = dl.GetDealInfo(dealID);

        // only consumer can blacklist
        require(msg.sender == consumerID || role == BlacklistPerson.BLACKLIST_NOBODY);
        if (role == BlacklistPerson.BLACKLIST_WORKER) {
            bl.Add(consumerID, supplierID);
        } else if (role == BlacklistPerson.BLACKLIST_MASTER) {
            bl.Add(consumerID, masterID);
        }
    }

    function InternalCloseDeal(uint dealID) internal {
        ( , address supplierID, address consumerID, address masterID, , , ) = dl.GetDealInfo(dealID);

        Deals.DealStatus status = dl.GetDealStatus(dealID);

        if (status == Deals.DealStatus.STATUS_CLOSED) {
            return;
        }
        require((status == Deals.DealStatus.STATUS_ACCEPTED));
        require(msg.sender == consumerID || msg.sender == supplierID || msg.sender == masterID);
        dl.SetDealStatus(dealID, Deals.DealStatus.STATUS_CLOSED);
        dl.SetDealEndTime(dealID, block.timestamp);
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
        pr = ProfileRegistry(_newPR);
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
        dl = Deals(_newDeals);
        return true;
    }

    function SetOrdersAddress(address _newOrders) public onlyOwner returns (bool) {
        ord = Orders(_newOrders);
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
        ord.transferOwnership(_newMarket);
        dl.transferOwnership(_newMarket);
        cr.transferOwnership(_newMarket);
        super.pause();

    }
}
