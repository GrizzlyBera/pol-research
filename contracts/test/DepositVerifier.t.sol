pragma solidity ^0.8.13;

import "../src/DepositVerifier.sol";
import { DepositVerifier } from "../src/DepositVerifier.sol";
import { Test } from "forge-std/Test.sol";
import { console } from "forge-std/console.sol";

contract NopDepositContract is IDepositContract {
    function deposit(
        bytes calldata pubkey,
        bytes calldata credentials,
        bytes calldata signature,
        address operator
    )
    external
    payable {}
}

contract DepositVerifierTest is Test {

    DepositVerifier public dv;

    bytes32 private currentDomain = 0x030000007a99249b35f928e89067829d3ddce74cf85cbe02b5c1aa54c2192beb;

    function setUp() public {
        NopDepositContract ndc = new NopDepositContract();
        dv = new DepositVerifier(address(ndc), currentDomain);
    }

    // deposit message hash root 0x707277169b083c3eb26e0fa59b7d321f59a835c2bedbbe91b831b183eff0b049

    // match result from beacon key CL
    function test_DepositRoot() public view {
        bytes memory pubKey = hex"af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46";
        bytes memory credentials = hex"010000000000000000000000036812189e22ddb7253f175d15ecbafa3c6b8ff4";
        uint amount = 10000000000000;
        bytes32 checkRoot = dv.computeDepositRoot(pubKey, credentials, amount);

        bytes32 expectedRoot = 0x707277169b083c3eb26e0fa59b7d321f59a835c2bedbbe91b831b183eff0b049;
        assertTrue(checkRoot == expectedRoot, "calculated deposit root should be same as expected");
    }

    // match result from beacon key CL
    function test_SigningRoot() public view {
        bytes memory pubKey = hex"af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46";
        bytes memory credentials = hex"010000000000000000000000036812189e22ddb7253f175d15ecbafa3c6b8ff4";
        uint amount = 10000000000000;
        bytes32 checkRoot = dv.computeSigningRoot(pubKey, credentials, amount);

        bytes32 expectedRoot = 0x33b70bfa533d21393c04e5858946c26f6c1ef29b9182f61442bd30a6386a3d64;
        assertTrue(checkRoot == expectedRoot, "calculated signing root should be same as expected");
    }

    function test_HashToCurve() public view {
        bytes32 testMsg = hex"deadbeef";

        BLS.G2Point memory g2A = dv.hashToCurve(testMsg);

        // derived from external test (TestBlstCompat) that uses blst HashToG2 to check for compat
        bytes memory expectSerialized = hex"000000000000000000000000000000000c41e5b6b7269ad84e0b7f84a38c4394c27ad84fb1b1297e7cc4a4d327a9579bc467b39a0730ead7e1047a76939fb1dc0000000000000000000000000000000012789f23883babd08d0dc6ffa786a5e75becdf9b243b00c301207085ec3f361c63bc9788b454869749d3b3ca459da5ea00000000000000000000000000000000079b97db530439bdf31b7b73081a4a66a99699bc5783dea78db421d7412f74c58710c4247f5082b2e8d830f00925527c0000000000000000000000000000000008647dc0d71669aae8a6498fd16090f9ca47562116d00755d15a37476125a3ba49cbf995120376dda25cff5309588492";
        require(expectSerialized.length==256);

        bytes memory gotSerialized = abi.encode(g2A);

        require(keccak256(expectSerialized)==keccak256(gotSerialized));
    }

    function test_PublicKeySerialization() public view {
        // this is the compressed public key - the form that is output by beacon kit
        bytes memory pubKeyCompressed = hex"af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46";
        // this is the encoded y coordinate of the public key. we will need beacon kit to output this as well
        bytes memory pubKeyY = hex"19738c95acc788168ffe2a2d35cf1b77514dece00f6e983dae8c09b1f3dc9f439ec48de6881f9544a17b04e4def6ede0";

        assert(pubKeyCompressed.length==48);
        BLS.G1Point memory decG1 = dv.decodeG1Point(pubKeyCompressed, pubKeyY);

        bytes memory encG1 = hex"000000000000000000000000000000000f7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a460000000000000000000000000000000019738c95acc788168ffe2a2d35cf1b77514dece00f6e983dae8c09b1f3dc9f439ec48de6881f9544a17b04e4def6ede0";

        assert(encG1.length==4*32);

        BLS.G1Point memory expectG1 = abi.decode(encG1, (BLS.G1Point));

        assert(decG1.x.a==expectG1.x.a);
        assert(decG1.x.b==expectG1.x.b);
        assert(decG1.y.a==expectG1.y.a);
        assert(decG1.y.b==expectG1.y.b);
    }

    function test_SignatureSerialization() public view {
        bytes memory signatureCompressed = hex"8e9bcb85781719ed1525d0e29a1865e8063d66c24c3ff3327178c0308107ea5810d317f64d4ad03bd43629ed82a0c1980bd46bc0468b20a991e7af28019e3fab66d0d27778bde8f530836b30c938ee8a73996d60fa79c0d5daf45c01b66b4a2d";
        bytes memory signatureY = hex"044ccf34a43bddaba88917788e37b1a160e464910f53feb80ce9f438061516b4c9633c7d1f651af4266412d1c33d814d0fd6c63ceab9a77120a35424fe30687c407aa9b6655c23303f7fa2135989c776e72f5730488e63e3b6b063e3f87c7f42";

        BLS.G2Point memory decG2 = dv.decodeG2Point(signatureCompressed, signatureY);

        bytes memory encG2 = hex"000000000000000000000000000000000bd46bc0468b20a991e7af28019e3fab66d0d27778bde8f530836b30c938ee8a73996d60fa79c0d5daf45c01b66b4a2d000000000000000000000000000000000e9bcb85781719ed1525d0e29a1865e8063d66c24c3ff3327178c0308107ea5810d317f64d4ad03bd43629ed82a0c198000000000000000000000000000000000fd6c63ceab9a77120a35424fe30687c407aa9b6655c23303f7fa2135989c776e72f5730488e63e3b6b063e3f87c7f4200000000000000000000000000000000044ccf34a43bddaba88917788e37b1a160e464910f53feb80ce9f438061516b4c9633c7d1f651af4266412d1c33d814d";
        assert(encG2.length==8*32);

        BLS.G2Point memory expectG2 = abi.decode(encG2, (BLS.G2Point));

        assert(decG2.x.c0.a==expectG2.x.c0.a);
        assert(decG2.x.c0.b==expectG2.x.c0.b);
        assert(decG2.x.c1.a==expectG2.x.c1.a);
        assert(decG2.x.c1.b==expectG2.x.c1.b);

        assert(decG2.y.c0.a==expectG2.y.c0.a);
        assert(decG2.y.c0.b==expectG2.y.c0.b);
        assert(decG2.y.c1.a==expectG2.y.c1.a);
        assert(decG2.y.c1.b==expectG2.y.c1.b);

    }

    function test_BLSSignatureVerify() public view {
        bytes32 expectedRoot = 0x33b70bfa533d21393c04e5858946c26f6c1ef29b9182f61442bd30a6386a3d64;

        bytes memory pubKeyCompressed = hex"af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46";
        bytes memory pubKeyY = hex"19738c95acc788168ffe2a2d35cf1b77514dece00f6e983dae8c09b1f3dc9f439ec48de6881f9544a17b04e4def6ede0";

        bytes memory signatureCompressed = hex"8e9bcb85781719ed1525d0e29a1865e8063d66c24c3ff3327178c0308107ea5810d317f64d4ad03bd43629ed82a0c1980bd46bc0468b20a991e7af28019e3fab66d0d27778bde8f530836b30c938ee8a73996d60fa79c0d5daf45c01b66b4a2d";
        bytes memory signatureY = hex"044ccf34a43bddaba88917788e37b1a160e464910f53feb80ce9f438061516b4c9633c7d1f651af4266412d1c33d814d0fd6c63ceab9a77120a35424fe30687c407aa9b6655c23303f7fa2135989c776e72f5730488e63e3b6b063e3f87c7f42";

        assert(dv.blsSignatureIsValid(expectedRoot, pubKeyCompressed, signatureCompressed, pubKeyY, signatureY));
    }

    function test_VerifyAndDeposit() public {

        bytes memory pubKeyCompressed = hex"af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46";
        bytes memory pubKeyY = hex"19738c95acc788168ffe2a2d35cf1b77514dece00f6e983dae8c09b1f3dc9f439ec48de6881f9544a17b04e4def6ede0";

        bytes memory credentials = hex"010000000000000000000000036812189e22ddb7253f175d15ecbafa3c6b8ff4";
        uint amount = 10000000000000;

        bytes memory signatureCompressed = hex"8e9bcb85781719ed1525d0e29a1865e8063d66c24c3ff3327178c0308107ea5810d317f64d4ad03bd43629ed82a0c1980bd46bc0468b20a991e7af28019e3fab66d0d27778bde8f530836b30c938ee8a73996d60fa79c0d5daf45c01b66b4a2d";
        bytes memory signatureY = hex"044ccf34a43bddaba88917788e37b1a160e464910f53feb80ce9f438061516b4c9633c7d1f651af4266412d1c33d814d0fd6c63ceab9a77120a35424fe30687c407aa9b6655c23303f7fa2135989c776e72f5730488e63e3b6b063e3f87c7f42";

        dv.verifyAndDeposit{value: amount}(pubKeyCompressed, credentials, signatureCompressed,
            address(0x00), pubKeyY, signatureY);

        signatureCompressed[0] = 0x8d;
        vm.expectRevert();
        dv.verifyAndDeposit{value: amount}(pubKeyCompressed, credentials, signatureCompressed,
            address(0x00), pubKeyY, signatureY);
    }
}