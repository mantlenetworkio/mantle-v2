pragma solidity ^0.8.9;

interface IERC165 {
    /// @notice Query if a contract implements an interface
    /// @param interfaceID The interface identifier, as specified in ERC-165
    /// @dev Interface identification is specified in ERC-165. This function
    ///  uses less than 30,000 gas.
    /// @return `true` if the contract implements `interfaceID` and
    ///  `interfaceID` is not 0xffffffff, `false` otherwise
    function supportsInterface(bytes4 interfaceID) external view returns (bool);
}

contract ERC165 is IERC165{
    bytes4 private constant _INTERFACE_ID_ERC165 = 0x01ffc9a7;

    mapping(bytes4=>bool) private _supportInterface;

    constructor() public {
        registerInterface(_INTERFACE_ID_ERC165);
    }

    function supportsInterface(bytes4 interfaceID) external override view returns (bool){
        return _supportInterface[interfaceID];
    }

    function registerInterface(bytes4 interfaceID) public {
        require(interfaceID != 0xffffffff);
        _supportInterface[interfaceID] = true;
    }
}
