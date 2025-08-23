package crypto

import (
	"encoding/hex"
	blst "github.com/supranational/blst/bindings/go"
	"testing"
)

// serilzied as xA1 | xA0 | yA1 | yA0
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
	println(hex.EncodeToString(ethCompatSerial))

	// g2Bytes := encodePointG2(g2.ToAffine())
	// println(hex.EncodeToString(g2Bytes))

	//  16292894393230995312347007029565211540
	//  87965740892816491579742703639908851599849623885422761044204589994416038457820
	//  24552407265427432973356939275716175335
	//  41578988400407522664178934841715549317899658127572562258297829970299894736362
	//  10112482004077415616164809005373999718
	//  76706959515166227106522159834704144135181742223203261725305249196730940215932
	//  11155604231706841424451390488031826169
	//  91493235997685588184210951294669436259063052349741381913249493266346108552338

	//checkH, _ := new(big.Int).SetString("10112482004077415616164809005373999718", 10)
	//checkL, _ := new(big.Int).SetString("76706959515166227106522159834704144135181742223203261725305249196730940215932", 10)
	//check := append(checkH.Bytes(), checkL.Bytes()...)
	//println(hex.EncodeToString(check))

}
