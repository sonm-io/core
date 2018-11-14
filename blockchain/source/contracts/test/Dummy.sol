pragma solidity ^0.4.23;

import "../Administratable.sol";

contract Dummy is Administratable {

    event Call(string method);

    function testOnlyOwner() public onlyOwner {
        emit Call("onlyOwner");
    }

    function testPublic() public returns (bool) {
        emit Call("public");
    }

    function testOwnerOrAdministrator() public ownerOrAdministrator returns (bool) {
        emit Call("ownerOrAdministrator");
    }
}
