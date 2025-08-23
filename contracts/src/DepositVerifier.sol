// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.10;

import { BLS } from "./BLS.sol";

contract DepositVerifier {

    // uint constant WEI_PER_GWEI = 1e9;

    // IDepositContract immutable depositContract;
    // Constant related to versioning serializations of deposits on eth2
    bytes32 immutable DEPOSIT_DOMAIN;

    constructor(address /* depositContractAddress */, bytes32 deposit_domain) {
        // depositContract = IDepositContract(depositContractAddress);
        DEPOSIT_DOMAIN = deposit_domain;
    }

    // cribbed: https://github.com/ethereum/consensus-specs/blob/62d253d6262eb7e6ea00f902466fd8002833b93b/solidity_deposit_contract/deposit_contract.sol#L165
    function to_little_endian_64(uint64 value) internal pure returns (bytes memory ret) {
        ret = new bytes(8);
        bytes8 bytesValue = bytes8(value);
        // Byteswapping during copying to bytes.
        ret[0] = bytesValue[7];
        ret[1] = bytesValue[6];
        ret[2] = bytesValue[5];
        ret[3] = bytesValue[4];
        ret[4] = bytesValue[3];
        ret[5] = bytesValue[2];
        ret[6] = bytesValue[1];
        ret[7] = bytesValue[0];
    }

    // TODO: this is just hand-rolled SSZ serialization and hashing...  library?
    function computeDepositRoot(
        bytes memory publicKey,
        bytes memory withdrawalCredentials,
        uint depositAmount
    ) public pure returns (bytes32) {
        bytes32 publicKeyRoot = sha256(abi.encodePacked(publicKey, bytes16(0)));
        bytes32 firstNode = sha256(abi.encodePacked(publicKeyRoot, withdrawalCredentials));

        bytes memory amount = to_little_endian_64(uint64(depositAmount));
        bytes32 secondNode = sha256(abi.encodePacked(amount, bytes24(0), bytes32(0)));

        return sha256(abi.encodePacked(firstNode, secondNode));
    }

    function computeSigningRoot(
        bytes memory publicKey,
        bytes memory withdrawalCredentials,
        uint amount
    ) public view returns (bytes32) {
        bytes32 depositMessageRoot = computeDepositRoot(publicKey, withdrawalCredentials, amount);
        return sha256(abi.encodePacked(depositMessageRoot, DEPOSIT_DOMAIN));
    }

    // Implements "hash to the curve" from the IETF BLS draft.
    function hashToCurve(bytes32 message) public view returns (BLS.G2Point memory) {
        return BLS.hashToCurveG2(bytes.concat(message));
    }
}
