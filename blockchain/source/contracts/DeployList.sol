pragma solidity ^0.4.23;


import "zeppelin-solidity/contracts/ownership/Ownable.sol";


contract DeployList is Ownable {
    address[] deployers;

    event DeployerAdded(address deployer);

    event DeployerRemoved(address deployer);

    constructor(address[] _deployers){
        owner = msg.sender;
        deployers = _deployers;
    }

    function addDeployer(address _deployer) public onlyOwner {
        deployers.push(_deployer);
        emit DeployerAdded(_deployer);
    }

    function removeDeployer(address _deployer) public onlyOwner {
        for (uint i = 0; i < deployers.length - 1; i++)
            if (deployers[i] == _deployer) {
                deployers[i] = deployers[deployers.length - 1];
                break;
            }
        emit DeployerRemoved(_deployer);
    }

    function getDeployers() public view returns (address[]){
        return deployers;
    }
}
