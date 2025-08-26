package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	blst "github.com/supranational/blst/bindings/go"
	"testing"
)

// serialized as xA1 | xA0 | yA1 | yA0
func ethEncodePointG2(g2 *blst.P2Affine) []byte {
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

	println(len(g2.Compress()))

	ethCompatSerial := ethEncodePointG2(g2.ToAffine())
	require.Equal(t,
		"000000000000000000000000000000000c41e5b6b7269ad84e0b7f84a38c4394c27ad84fb1b1297e7cc4a4d327a9579bc467b39a0730ead7e1047a76939fb1dc0000000000000000000000000000000012789f23883babd08d0dc6ffa786a5e75becdf9b243b00c301207085ec3f361c63bc9788b454869749d3b3ca459da5ea00000000000000000000000000000000079b97db530439bdf31b7b73081a4a66a99699bc5783dea78db421d7412f74c58710c4247f5082b2e8d830f00925527c0000000000000000000000000000000008647dc0d71669aae8a6498fd16090f9ca47562116d00755d15a37476125a3ba49cbf995120376dda25cff5309588492",
		hex.EncodeToString(ethCompatSerial))
}

func TestG1PointDeserialize(t *testing.T) {
	// af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46

	g1CompressedBytes, _ := hex.DecodeString("af7bc18c871e23775aba48f19964516922c5afb835ddf4b3edb951d115a5c89f6db5445f7f8aa04394054275fd4c4a46")

	p1 := new(blst.P1Affine)
	p1.Uncompress(g1CompressedBytes)

	p1SerialBytes := p1.Serialize()

	lpad := make([]byte, 16)

	println(hex.EncodeToString(append(lpad, p1SerialBytes[0:16]...)))
	println(hex.EncodeToString(p1SerialBytes[16:48]))
	println(hex.EncodeToString(append(lpad, p1SerialBytes[48:64]...)))
	println(hex.EncodeToString(p1SerialBytes[64:96]))

	println()

	// this is the 'encoded y' component
	println(hex.EncodeToString(p1SerialBytes[48:]))

}

func TestG2PointDeserialize(t *testing.T) {
	g2CompressedBytes, _ := hex.DecodeString("8e9bcb85781719ed1525d0e29a1865e8063d66c24c3ff3327178c0308107ea5810d317f64d4ad03bd43629ed82a0c1980bd46bc0468b20a991e7af28019e3fab66d0d27778bde8f530836b30c938ee8a73996d60fa79c0d5daf45c01b66b4a2d")

	p2 := new(blst.P2Affine)
	p2.Uncompress(g2CompressedBytes)

	g2SerialBytes := p2.Serialize()
	require.Equal(t, len(g2SerialBytes), 192)

	// this is the 'encoded y' component
	println(hex.EncodeToString(g2SerialBytes[96:]))

	ethEncodedG2 := ethEncodePointG2(p2)
	println(hex.EncodeToString(ethEncodedG2))

}
