// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.10;

import { BLS } from "./BLS.sol";

contract DepositVerifier {

    // uint constant WEI_PER_GWEI = 1e9;

    // IDepositContract immutable depositContract;
    // Constant related to versioning serializations of deposits on eth2
    bytes32 immutable DEPOSIT_DOMAIN;

    uint8 constant MOD_EXP_PRECOMPILE_ADDRESS = 0x5;
    uint constant BLS12_381_G2_ADD_ADDRESS = 0x0d;
    uint8 constant BLS12_381_MAP_FIELD_TO_CURVE_PRECOMPILE_ADDRESS = 0x11;

    // Fp is a field element with the high-order part stored in `a`.
    struct Fp {
        uint a;
        uint b;
    }

    // Fp2 is an extension field element with the coefficient of the
    // quadratic non-residue stored in `b`, i.e. p = a + i * b
    struct Fp2 {
        Fp a;
        Fp b;
    }

    // G2Point represents a point on BLS12-381 over Fp2 with coordinates (X,Y);
    struct G2Point {
        Fp2 X;
        Fp2 Y;
    }

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

    string constant BLS_SIG_DST = "BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_NUL_";
    uint8 constant BLS_SIG_DST_LEN = 43;

    function expandMessage(bytes32 message) public pure returns (bytes memory) {

        // [hash block size = 64] + [msg size = 32] + 3 + [DST size = 43] + 1 = 143
        bytes32 b0 = sha256(abi.encodePacked(
            bytes32(0), bytes32(0), message, bytes1(0x01), bytes1(0x00), bytes1(0x00), BLS_SIG_DST, bytes1(BLS_SIG_DST_LEN)
        ));

        bytes memory output = new bytes(256);
        bytes32 chunk = sha256(abi.encodePacked(b0, bytes1(0x01), bytes(BLS_SIG_DST)));
        assembly {
            mstore(add(output, 0x20), chunk)
        }
        for (uint i = 2; i < 9; i++) {
            bytes32 input;
            assembly {
                input := xor(b0, mload(add(output, add(0x20, mul(0x20, sub(i, 2))))))
            }
            chunk = sha256(abi.encodePacked(input, bytes1(uint8(i)), bytes(BLS_SIG_DST)));
            assembly {
                mstore(add(output, add(0x20, mul(0x20, sub(i, 1)))), chunk)
            }
        }

        return output;
    }

    // TODO: abi encode?
    function sliceToUint(bytes memory data, uint start, uint end) private pure returns (uint) {
        uint length = end - start;
        assert(length >= 0);
        assert(length <= 32);

        uint result;
        for (uint i = 0; i < length; i++) {
            bytes1 b = data[start+i];
            result = result + (uint8(b) * 2**(8*(length-i-1)));
        }
        return result;
    }

    // much cribbing from https://github.com/ralexstokes/deposit-verifier/blob/084186bc58a2a03fcc3e1fc4263723655fa899a1/deposit_verifier.sol#L142C1-L198C1
    function reduceModulo(bytes memory data, uint start, uint end) private view returns (bytes memory) {
        uint length = end - start;
        assert (length >= 0);
        assert (length <= data.length);

        bytes memory result = new bytes(48);

        bool success;
        assembly {
            let p := mload(0x40)
        // length of base
            mstore(p, length)
        // length of exponent
            mstore(add(p, 0x20), 0x20)
        // length of modulus
            mstore(add(p, 0x40), 48)
        // base
        // first, copy slice by chunks of EVM words
            let ctr := length
            let src := add(add(data, 0x20), start)
            let dst := add(p, 0x60)
            for { }
            or(gt(ctr, 0x20), eq(ctr, 0x20))
            { ctr := sub(ctr, 0x20) }
            {
                mstore(dst, mload(src))
                dst := add(dst, 0x20)
                src := add(src, 0x20)
            }
        // next, copy remaining bytes in last partial word
            let mask := sub(exp(256, sub(0x20, ctr)), 1)
            let srcpart := and(mload(src), not(mask))
            let destpart := and(mload(dst), mask)
            mstore(dst, or(destpart, srcpart))
        // exponent
            mstore(add(p, add(0x60, length)), 1)
        // modulus
            let modulusAddr := add(p, add(0x60, add(0x10, length)))
            mstore(modulusAddr, or(mload(modulusAddr), 0x1a0111ea397fe69a4b1ba7b6434bacd7)) // pt 1
            mstore(add(p, add(0x90, length)), 0x64774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab) // pt 2
            success := staticcall(
                sub(gas(), 2000),
                MOD_EXP_PRECOMPILE_ADDRESS,
                p,
                add(0xB0, length),
                add(result, 0x20),
                48)
        // Use "invalid" to make gas estimation work
            switch success case 0 { invalid() }
        }
        require(success, "call to modular exponentiation precompile failed");
        return result;
    }

    function convertSliceToFp(bytes memory data, uint start, uint end) private view returns (Fp memory) {
        bytes memory fieldElement = reduceModulo(data, start, end);
        uint a = sliceToUint(fieldElement, 0, 16);
        uint b = sliceToUint(fieldElement, 16, 48);
        return Fp(a, b);
    }

    function hashToField(bytes32 message) public view returns (Fp2[2] memory result) {
        bytes memory some_bytes = expandMessage(message);
        result[0] = Fp2(
            convertSliceToFp(some_bytes, 0, 64),
            convertSliceToFp(some_bytes, 64, 128)
        );
        result[1] = Fp2(
            convertSliceToFp(some_bytes, 128, 192),
            convertSliceToFp(some_bytes, 192, 256)
        );
    }

    function mapToCurve(Fp2 memory fieldElement) public view returns (G2Point memory result) {
        uint[4] memory input;
        input[0] = fieldElement.a.a;
        input[1] = fieldElement.a.b;
        input[2] = fieldElement.b.a;
        input[3] = fieldElement.b.b;

        uint[8] memory output;

        bool success;
        assembly {
            success := staticcall(
                sub(gas(), 2000),
                BLS12_381_MAP_FIELD_TO_CURVE_PRECOMPILE_ADDRESS,
                input,
                128,
                output,
                256
            )
        // Use "invalid" to make gas estimation work
            switch success case 0 { invalid() }
        }
        require(success, "call to map to curve precompile failed");

        return G2Point(
            Fp2(
                Fp(output[0], output[1]),
                Fp(output[2], output[3])
            ),
            Fp2(
                Fp(output[4], output[5]),
                Fp(output[6], output[7])
            )
        );
    }

    function addG2(G2Point memory a, G2Point memory b) private view returns (G2Point memory) {
        uint[16] memory input;
        input[0]  = a.X.a.a;
        input[1]  = a.X.a.b;
        input[2]  = a.X.b.a;
        input[3]  = a.X.b.b;
        input[4]  = a.Y.a.a;
        input[5]  = a.Y.a.b;
        input[6]  = a.Y.b.a;
        input[7]  = a.Y.b.b;

        input[8]  = b.X.a.a;
        input[9]  = b.X.a.b;
        input[10] = b.X.b.a;
        input[11] = b.X.b.b;
        input[12] = b.Y.a.a;
        input[13] = b.Y.a.b;
        input[14] = b.Y.b.a;
        input[15] = b.Y.b.b;

        uint[8] memory output;

        bool success;
        assembly {
            success := staticcall(
                sub(gas(), 2000),
                BLS12_381_G2_ADD_ADDRESS,
                input,
                512,
                output,
                256
            )
        // Use "invalid" to make gas estimation work
            switch success case 0 { invalid() }
        }
        require(success, "call to addition in G2 precompile failed");

        return G2Point(
            Fp2(
                Fp(output[0], output[1]),
                Fp(output[2], output[3])
            ),
            Fp2(
                Fp(output[4], output[5]),
                Fp(output[6], output[7])
            )
        );
    }

    // Implements "hash to the curve" from the IETF BLS draft.
    function hashToCurve(bytes32 message) public view returns (G2Point memory) {
        // return BLS.hashToCurveG2(bytes.concat(message));
        Fp2[2] memory messageElementsInField = hashToField(message);
        G2Point memory firstPoint = mapToCurve(messageElementsInField[0]);
        G2Point memory secondPoint = mapToCurve(messageElementsInField[1]);
        return addG2(firstPoint, secondPoint);
    }
}
