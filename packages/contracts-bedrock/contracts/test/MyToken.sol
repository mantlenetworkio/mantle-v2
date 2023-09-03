pragma solidity ^0.8.7;

import "./IERC721.sol";
import "./ERC165.sol";
import "./SafeMath.sol";

interface IERC721TokenReceiver {
    function onERC721Received(address _operator, address _from, uint256 _tokenId, bytes calldata _data) external returns(bytes4);
}

contract ERC721 is ERC165, IERC721 {

    string public name     = "MyToken";

    string public symbol   = "MT";

    mapping(address => uint256) private _ownerTokensCount;

    mapping(uint256 => address) private _tokenOwner;

    mapping(uint256 => address) private _tokenApproval;

    mapping(address => mapping(address => bool)) private _operatorApprovals;

    using SafeMath for uint256;

    bytes4 private constant _INTERFACE_ID_ERC721 = 0x80ac58cd;

    bytes4 private constant _ERC721_RECEIVED = 0x150b7a02;

    constructor() public {
        registerInterface(_INTERFACE_ID_ERC721);
    }

    modifier ownerIsNotZeroAddress(address _owner) {
        require(address(0) != _owner);
        _;
    }

    function balanceOf(address _owner) ownerIsNotZeroAddress(_owner) external override view returns (uint256) {
        return _ownerTokensCount[_owner];
    }

    function ownerOf(uint256 _tokenId) external override view returns (address) {
        return _ownerOf(_tokenId);
    }

    function _ownerOf(uint256 _tokenId) internal view returns (address) {
        address _owner = _tokenOwner[_tokenId];
        require(address(0) != _owner);

        return _owner;
    }

    function _transferFrom(address _from, address _to, uint256 _tokenId) internal virtual {
        require(_ownerOf(_tokenId) == _from);
        require(address(0) != _to);
        require(_isApprovedOrOwner(msg.sender, _tokenId));
        _approve(address(0), _tokenId);
        _ownerTokensCount[_from] = _ownerTokensCount[_from].sub(1);
        _ownerTokensCount[_to] = _ownerTokensCount[_to].add(1);
        _tokenOwner[_tokenId] = _to;

        emit Transfer(_from, _to, _tokenId);
    }

    function _isExistTokenId(uint256 _tokenId) internal view returns (bool) {
        address owner = _tokenOwner[_tokenId];
        if (address(0) != owner) {
            return true;
        }
        return false;
    }

    function _isApprovedOrOwner(address _spender, uint256 _tokenId) internal view returns (bool) {
        require(_isExistTokenId(_tokenId));
        address _owner = _ownerOf(_tokenId);

        return (_owner == _spender || _getApproved(_tokenId) == _spender || _isApprovedForAll(_owner, _spender));
    }

    function _isContract(address addr) internal view returns (bool) {
        uint256 _size;

        assembly { _size := extcodesize(addr) }
        return _size > 0;
    }

    function _checkOnERC721Received(address _from, address _to, uint256 _tokenId, bytes memory _data) private returns (bool) {
        if (!_isContract(_to)) {
            return true;
        }

        (bool success, bytes memory returndata) = _to.call(abi.encodeWithSelector(
            IERC721TokenReceiver(_to).onERC721Received.selector,
            msg.sender,
            _from,
            _tokenId,
            _data
        ));

        if (!success) {
            revert();
        } else {
            bytes4 retval = abi.decode(returndata, (bytes4));
            return (retval == _ERC721_RECEIVED);
        }
    }

    function _safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes memory _data) internal virtual {
        _transferFrom(_from, _to, _tokenId);
        require(_checkOnERC721Received(_from, _to, _tokenId, _data));
    }

    function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes calldata data) external override payable {
        _safeTransferFrom(_from, _to, _tokenId, data);
    }

    function safeTransferFrom(address _from, address _to, uint256 _tokenId) external override payable {
        _safeTransferFrom(_from, _to, _tokenId, "");
    }

    function transferFrom(address _from, address _to, uint256 _tokenId) external override payable {
        _transferFrom(_from, _to, _tokenId);
    }

    function _approve(address _approved, uint256 _tokenId)  internal  {
        require(_approved != msg.sender);
        _tokenApproval[_tokenId] = _approved;
    }

    function approve(address _approved, uint256 _tokenId) ownerIsNotZeroAddress(_approved) external override payable {
        _approve(_approved, _tokenId);
    }

    function setApprovalForAll(address _operator, bool _approved) external override {
        require(_operator != msg.sender);
        _operatorApprovals[msg.sender][_operator] = _approved;

        emit ApprovalForAll(msg.sender, _operator, _approved);
    }

    function _getApproved(uint256 _tokenId) internal view returns (address) {
        require(_isExistTokenId(_tokenId));
        return _tokenApproval[_tokenId];
    }

    function getApproved(uint256 _tokenId) external override view returns (address) {
        return _getApproved(_tokenId);
    }

    function _isApprovedForAll(address _owner, address _operator) internal view returns (bool) {
        return _operatorApprovals[_owner][_operator];
    }

    function isApprovedForAll(address _owner, address _operator) external override view returns (bool) {
        return _isApprovedForAll(_owner, _operator);
    }

    function _mint(address _to, uint256 _tokenId) ownerIsNotZeroAddress(_to) internal virtual {
        require(!_isExistTokenId(_tokenId));
        _tokenOwner[_tokenId] = _to;
        _ownerTokensCount[_to] = _ownerTokensCount[_to].add(1);

        emit Transfer(address(0), _to, _tokenId);
    }

    function mint(address _to, uint256 _tokenId) ownerIsNotZeroAddress(_to) external {
        _mint(_to, _tokenId);
    }

    function safeMint(address _to, uint256 _tokenId, bytes calldata _data) ownerIsNotZeroAddress(_to) external {
        _mint(_to, _tokenId);
        require(_checkOnERC721Received(address(0), _to, _tokenId, _data));
    }

}
