pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";
import "zeppelin-solidity/contracts/math/SafeMath.sol";
import "./Orders.sol";


contract ChangeRequests is Ownable {
    using SafeMath for uint256;

    mapping(uint => ChangeRequest) requests;

    mapping(uint => uint[2]) actualRequests;

    uint requestsAmount;

    struct ChangeRequest {
        uint dealID;
        Orders.OrderType requestType;
        uint price;
        uint duration;
        RequestStatus status;
    }

    enum RequestStatus {
        REQUEST_UNKNOWN,
        REQUEST_CREATED,
        REQUEST_CANCELED,
        REQUEST_REJECTED,
        REQUEST_ACCEPTED
    }


    constructor() public {
        owner = msg.sender;
    }

    function Write(
        uint _dealID,
        Orders.OrderType _requestType,
        uint _price,
        uint _duration,
        RequestStatus _status) public onlyOwner returns(uint){

        requestsAmount = requestsAmount.add(1);

        requests[requestsAmount] = ChangeRequest(_dealID, _requestType, _price, _duration, _status);

        return  requestsAmount;
    }

    //SETTERS

    function SetChangeRequestDealID(uint _changeRequestID, uint _dealID) public onlyOwner {
        requests[_changeRequestID].dealID = _dealID;
    }

    function SetChangeRequestType(uint _changeRequestID, Orders.OrderType _type) public onlyOwner {
        requests[_changeRequestID].requestType = _type;
    }

    function SetChangeRequestPrice(uint _changeRequestID, uint _price) public onlyOwner {
        requests[_changeRequestID].price = _price;
    }

    function SetChangeRequestDuration(uint _changeRequestID, uint _duration) public onlyOwner {
        requests[_changeRequestID].duration = _duration;
    }

    function SetChangeRequestStatus(uint _changeRequestID, RequestStatus _status) public onlyOwner {
        requests[_changeRequestID].status = _status;
    }

    function SetActualChangeRequest(uint dealID, uint role, uint _changeRequestID) public onlyOwner {
        actualRequests[dealID][role] = _changeRequestID;
    }

    // GETTERS

    function GetChangeRequestDealID(uint _changeRequestID) public view returns(uint) {
        return requests[_changeRequestID].dealID;
    }

    function GetChangeRequestType(uint _changeRequestID) public view returns(Orders.OrderType) {
        return requests[_changeRequestID].requestType;
    }

    function GetChangeRequestPrice(uint _changeRequestID) public view returns(uint) {
        return requests[_changeRequestID].price;
    }

    function GetChangeRequestDuration(uint _changeRequestID) public view returns(uint) {
        return requests[_changeRequestID].duration;
    }

    function GetChangeRequestStatus(uint _changeRequestID) public view returns(RequestStatus) {
        return requests[_changeRequestID].status;
    }

    function GetActualChangeRequest(uint dealID, uint role)public view returns(uint) {
        return actualRequests[dealID][role];

    }

    function GetChangeRequestsAmount() public view returns(uint)  {
        return requestsAmount;
    }

    function GetChangeRequestInfo(uint changeRequestID) public view
    returns (
        uint dealID,
        Orders.OrderType requestType,
        uint price,
        uint duration,
        RequestStatus status
    ) {
        return (
        requests[changeRequestID].dealID,
        requests[changeRequestID].requestType,
        requests[changeRequestID].price,
        requests[changeRequestID].duration,
        requests[changeRequestID].status
        );
    }


}
