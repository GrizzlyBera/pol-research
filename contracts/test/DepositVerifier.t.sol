pragma solidity ^0.8.13;

import "../src/DepositVerifier.sol";
import { DepositVerifier } from "../src/DepositVerifier.sol";
import { Test } from "forge-std/Test.sol";
import { console } from "forge-std/console.sol";

contract DepositVerifierTest is Test {

    DepositVerifier public dv;

    bytes32 private currentDomain = 0x030000007a99249b35f928e89067829d3ddce74cf85cbe02b5c1aa54c2192beb;

    function setUp() public {
        dv = new DepositVerifier(address(0x0), currentDomain);
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
        bytes memory bytesMessage = bytes.concat(testMsg);
        require(bytesMessage.length==32);

        BLS.G2Point memory g2A = BLS.hashToCurveG2(bytesMessage);

        // derived from external test that uses blst HashToG2 to check for compat
        bytes memory expectSerialized = hex"000000000000000000000000000000000c41e5b6b7269ad84e0b7f84a38c4394c27ad84fb1b1297e7cc4a4d327a9579bc467b39a0730ead7e1047a76939fb1dc0000000000000000000000000000000012789f23883babd08d0dc6ffa786a5e75becdf9b243b00c301207085ec3f361c63bc9788b454869749d3b3ca459da5ea00000000000000000000000000000000079b97db530439bdf31b7b73081a4a66a99699bc5783dea78db421d7412f74c58710c4247f5082b2e8d830f00925527c0000000000000000000000000000000008647dc0d71669aae8a6498fd16090f9ca47562116d00755d15a37476125a3ba49cbf995120376dda25cff5309588492";
        require(expectSerialized.length==256);

        bytes memory gotSerialized = abi.encode(g2A);

        require(keccak256(expectSerialized)==keccak256(gotSerialized));
    }
}