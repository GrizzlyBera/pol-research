// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.10;

import { BLS } from "./BLS.sol";

// TODO
interface IDepositContract {
    function deposit(
        bytes calldata pubkey,
        bytes calldata credentials,
        bytes calldata signature,
        address operator
    )
    external
    payable;
}

contract DepositVerifier {

    uint constant PUBLIC_KEY_LENGTH = 48;
    uint constant SIGNATURE_LENGTH = 96;
    uint constant WITHDRAWAL_CREDENTIALS_LENGTH = 32;

    // uint constant WEI_PER_GWEI = 1e9;

    bytes1 constant BLS_BYTE_WITHOUT_FLAGS_MASK = bytes1(0x1f);
    uint constant BLS_FIELD_ELEM_BYTE_LENGTH = 48;

    IDepositContract immutable private depositContract;

    // Constant related to versioning serializations of deposits on eth2
    // TODO: this changes (right?)
    bytes32 immutable DEPOSIT_DOMAIN;

    /// @notice The negated generator point in G1 (-P1). Used during pairing as a first G1 point.
    BLS.G1Point NEGATED_G1_GENERATOR = BLS.G1Point(
        BLS.Fp(
            31827880280837800241567138048534752271,
            88385725958748408079899006800036250932223001591707578097800747617502997169851
        ),
        BLS.Fp(
            22997279242622214937712647648895181298,
            46816884707101390882112958134453447585552332943769894357249934112654335001290
        )
    );

    constructor(address depositContractAddress, bytes32 deposit_domain) {
        depositContract = IDepositContract(depositContractAddress);
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

    // TODO: better way to do this? abi encoding?
    function sliceToUint(bytes memory data, uint start, uint end) private pure returns (uint256) {
        uint length = end - start;
        assert(length >= 0);
        assert(length <= 32);

        uint256 result;
        for (uint i = 0; i < length; i++) {
            bytes1 b = data[start+i];
            result = result + (uint8(b) * 2**(8*(length-i-1)));
        }
        return result;
    }

    function decodeG1Point(bytes memory encodedX, bytes memory encodedY) public pure returns (BLS.G1Point memory) {
        assert(encodedX.length==BLS_FIELD_ELEM_BYTE_LENGTH);
        assert(encodedY.length==BLS_FIELD_ELEM_BYTE_LENGTH);
        encodedX[0] = encodedX[0] & BLS_BYTE_WITHOUT_FLAGS_MASK;
        BLS.Fp memory X = BLS.Fp(sliceToUint(encodedX, 0, 16),
            sliceToUint(encodedX, 16, 48));
        BLS.Fp memory Y = BLS.Fp(sliceToUint(encodedY, 0, 16),
            sliceToUint(encodedY, 16, 48));
        return BLS.G1Point(X,Y);
    }

    function decodeFp2(bytes memory enc) private pure returns (BLS.Fp2 memory) {
        // NOTE: order is important here for decoding point...
        uint aa = sliceToUint(enc, 48, 64);
        uint ab = sliceToUint(enc, 64, 96);
        uint ba = sliceToUint(enc, 0, 16);
        uint bb = sliceToUint(enc, 16, 48);
        return BLS.Fp2(
            BLS.Fp(aa, ab),
            BLS.Fp(ba, bb)
        );
    }

    function decodeG2Point(bytes memory encodedX, bytes memory encodedY) public pure returns (BLS.G2Point memory) {
        encodedX[0] = encodedX[0] & BLS_BYTE_WITHOUT_FLAGS_MASK;
        // NOTE: the "flag bits" of the second half of `encodedX` are always == 0x0
        BLS.Fp2 memory X = decodeFp2(encodedX);
        BLS.Fp2 memory Y = decodeFp2(encodedY);
        return BLS.G2Point(X, Y);
    }

    function blsSignatureIsValid(
        bytes32 message,
        bytes memory encodedPublicKey,
        bytes memory encodedSignature,
        bytes memory publicKeyYCoordinate,
        bytes memory signatureYCoordinate
    ) public view returns (bool) {
        BLS.G1Point memory publicKey = decodeG1Point(encodedPublicKey, publicKeyYCoordinate);
        BLS.G2Point memory signature = decodeG2Point(encodedSignature, signatureYCoordinate);
        BLS.G2Point memory messageOnCurve = hashToCurve(message);

        BLS.G1Point[] memory g1Points = new BLS.G1Point[](2);
        BLS.G2Point[] memory g2Points = new BLS.G2Point[](2);

        g1Points[0] = NEGATED_G1_GENERATOR;
        g1Points[1] = publicKey;

        g2Points[0] = signature;
        g2Points[1] = messageOnCurve;

        // verify signature
        return BLS.Pairing(g1Points, g2Points);
    }

    // 1,040,464,384
    function verifyAndDeposit(
        bytes calldata publicKey,
        bytes calldata withdrawalCredentials,
        bytes calldata signature,
        address operator,
        bytes calldata publicKeyYCoordinate,
        bytes calldata signatureYCoordinate
    ) external payable {
        require(publicKey.length == PUBLIC_KEY_LENGTH, "incorrectly sized public key");
        require(withdrawalCredentials.length == WITHDRAWAL_CREDENTIALS_LENGTH, "incorrectly sized withdrawal credentials");
        require(signature.length == SIGNATURE_LENGTH, "incorrectly sized signature");

        // TODO: validate length of y coordinates?

        bytes32 signingRoot = computeSigningRoot(
            publicKey,
            withdrawalCredentials,
            msg.value
        );

        require(
            blsSignatureIsValid(
                signingRoot,
                publicKey,
                signature,
                publicKeyYCoordinate,
                signatureYCoordinate
            ),
            "BLS signature verification failed"
        );

        depositContract.deposit{value: msg.value}(publicKey, withdrawalCredentials, signature, operator);
    }
}
