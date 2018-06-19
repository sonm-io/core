pragma solidity ^0.4.23;

import "zeppelin-solidity/contracts/ownership/Ownable.sol";

contract ProfileRegistry is Ownable {

    modifier onlySonm(){
        require(GetValidatorLevel(msg.sender) == - 1);
        _;
    }

    enum IdentityLevel {
        UNKNOWN,
        ANONYMOUS,
        REGISTERED,
        IDENTIFIED,
        PROFESSIONAL
    }

    struct Certificate {
        address from;

        address to;

        uint attributeType;

        bytes value;
    }

    event ValidatorCreated(address indexed validator);

    event ValidatorDeleted(address indexed validator);

    event CertificateCreated(uint indexed id);

    event CertificateUpdated(uint indexed id);

    event ProfileRegistryHasBeenFrozen();

    bool isContractFrozen = false;

    uint256 certificatesCount = 0;

    mapping(address => int8) public validators;

    mapping(uint256 => Certificate) public certificates;

    mapping(address => mapping(uint256 => bytes)) certificateValue;

    mapping(address => mapping(uint256 => uint)) certificateCount;

    constructor() public {
        validators[msg.sender] = - 1;
    }

    function AddValidator(address _validator, int8 _level) onlySonm unfreezed public returns (address){
        require(_level > 0);
        require(GetValidatorLevel(_validator) == 0);
        validators[_validator] = _level;
        emit ValidatorCreated(_validator);
        return _validator;
    }

    function RemoveValidator(address _validator) onlySonm unfreezed public returns (address){
        require(GetValidatorLevel(_validator) > 0);
        validators[_validator] = 0;
        emit ValidatorDeleted(_validator);
        return _validator;
    }

    function GetValidatorLevel(address _validator) view public returns (int8){
        return validators[_validator];
    }

    function CreateCertificate(address _owner, uint256 _type, bytes _value) unfreezed public {
        //Check validator level
        if (_type >= 1100) {
            int8 attributeLevel = int8(_type / 100 % 10);
            require(attributeLevel <= GetValidatorLevel(msg.sender));
        } else {
            require(_owner == msg.sender);
        }

        // Check empty value
        require(keccak256(_value) != keccak256(""));

        bool isMultiple = _type / 1000 == 2;
        if (!isMultiple) {
            if (certificateCount[_owner][_type] == 0) {
                certificateValue[_owner][_type] = _value;
            } else {
                require(keccak256(GetAttributeValue(_owner, _type)) == keccak256(_value));
            }
        }

        certificateCount[_owner][_type] = certificateCount[_owner][_type] + 1;

        certificatesCount = certificatesCount + 1;
        certificates[certificatesCount] = Certificate(msg.sender, _owner, _type, _value);

        emit CertificateCreated(certificatesCount);
    }

    function RemoveCertificate(uint256 _id) unfreezed public {
        Certificate memory crt = certificates[_id];

        require(crt.to == msg.sender || crt.from == msg.sender || GetValidatorLevel(msg.sender) == -1);
        require(keccak256(crt.value) != keccak256(""));

        certificateCount[crt.to][crt.attributeType] = certificateCount[crt.to][crt.attributeType] - 1;
        if (certificateCount[crt.to][crt.attributeType] == 0) {
            certificateValue[crt.to][crt.attributeType] = "";
        }
        certificates[_id].value = "";
        emit CertificateUpdated(_id);
    }

    function GetCertificate(uint256 _id) view public returns (address, address, uint256, bytes){
        return (certificates[_id].from, certificates[_id].to, certificates[_id].attributeType, certificates[_id].value);
    }

    function GetAttributeValue(address _owner, uint256 _type) view public returns (bytes){
        return certificateValue[_owner][_type];
    }

    function GetAttributeCount(address _owner, uint256 _type) view public returns (uint256){
        return certificateCount[_owner][_type];
    }

    function GetProfileLevel(address _owner) view public returns (IdentityLevel){
        if (GetAttributeValue(_owner, 1401).length > 0) {
            return IdentityLevel.PROFESSIONAL;
        } else if (GetAttributeValue(_owner, 1301).length > 0) {
            return IdentityLevel.IDENTIFIED;
        } else if (GetAttributeValue(_owner, 1201).length > 0) {
            return IdentityLevel.REGISTERED;
        } else {
            return IdentityLevel.ANONYMOUS;
        }
    }

    function AddSonmRights(address _newSonm) onlyOwner public returns (bool) {
        validators[_newSonm] = -1;
        return true;
    }

    function RemoveSonmRights(address _oldSonm) onlyOwner public returns (bool) {
        require(GetValidatorLevel(_oldSonm) == -1);
        validators[_oldSonm] = 0;
        return true;
    }

    function FreezeProfileRegistry() onlyOwner public returns (bool) {
        isContractFrozen = true;
        emit ProfileRegistryHasBeenFrozen();
        return true;
    }
    //MODIFIERS

    modifier unfreezed() {
        require(isContractFrozen == false);
        _;
    }
}
