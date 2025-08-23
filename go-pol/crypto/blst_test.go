package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	blst "github.com/supranational/blst/bindings/go"
	"testing"
)

// serialized as xA1 | xA0 | yA1 | yA0
func ethEncodePointG2(g2 *blst.P2) []byte {
	out := make([]byte, 256)

	g2Bytes := g2.Serialize()

	xA1Bytes := g2Bytes[:48]
	xA0Bytes := g2Bytes[48:96]
	yA1Bytes := g2Bytes[96:144]
	yA0Bytes := g2Bytes[144:]

	copy(out[16:16+48], xA0Bytes[:])
	copy(out[80:80+48], xA1Bytes[:])
	copy(out[144:144+48], yA0Bytes[:])
	copy(out[208:208+48], yA1Bytes[:])

	return out
}

func TestBlstCompat(t *testing.T) {
	testMessage, _ := hex.DecodeString("deadbeef")
	messageBytes := make([]byte, 32)
	copy(messageBytes, testMessage)

	// REMINDER! blst HashToG2 is actually hash *and* map
	g2 := blst.HashToG2(messageBytes, []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_NUL_"))
	g2.Print("TEST")

	ethCompatSerial := ethEncodePointG2(g2)
	require.Equal(t,
		"000000000000000000000000000000000c41e5b6b7269ad84e0b7f84a38c4394c27ad84fb1b1297e7cc4a4d327a9579bc467b39a0730ead7e1047a76939fb1dc0000000000000000000000000000000012789f23883babd08d0dc6ffa786a5e75becdf9b243b00c301207085ec3f361c63bc9788b454869749d3b3ca459da5ea00000000000000000000000000000000079b97db530439bdf31b7b73081a4a66a99699bc5783dea78db421d7412f74c58710c4247f5082b2e8d830f00925527c0000000000000000000000000000000008647dc0d71669aae8a6498fd16090f9ca47562116d00755d15a37476125a3ba49cbf995120376dda25cff5309588492",
		hex.EncodeToString(ethCompatSerial))
}
