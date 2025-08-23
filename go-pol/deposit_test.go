package go_pol

import (
	"encoding/hex"
	"testing"
)

func TestDepositContractBasics(t *testing.T) {
	// 3,0,0,0,122,153,36,155,53,249,40,232,144,103,130,157,61,220,231,76,248,92,190,2,181,193,170,84,194,25,43,235
	currentDomainBytes := []byte{3, 0, 0, 0, 122, 153, 36, 155, 53, 249, 40, 232, 144, 103, 130, 157, 61, 220, 231, 76, 248, 92, 190, 2, 181, 193, 170, 84, 194, 25, 43, 235}
	currentDomainHex := hex.EncodeToString(currentDomainBytes)
	println(currentDomainHex)

	signingRootBytes := []byte{51, 183, 11, 250, 83, 61, 33, 57, 60, 4, 229, 133, 137, 70, 194, 111, 108, 30, 242, 155, 145, 130, 246, 20, 66, 189, 48, 166, 56, 106, 61, 100}
	signingRootHex := hex.EncodeToString(signingRootBytes)
	println(signingRootHex)
}
