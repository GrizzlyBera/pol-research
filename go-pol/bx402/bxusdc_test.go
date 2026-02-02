package bx402

import (
	"crypto/ecdsa"
	"github.com/GrizzlyBera/pol-research/go-pol/chain"
	"github.com/GrizzlyBera/pol-research/go-pol/crypto"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"os"
	"testing"
)

var bepoliaMasterMinter = ethgo.HexToAddress("0x5132D83Cb40367106BF3Ad2b845E521bf82CFc6a")
var bepoliaFiatTokenProxy = ethgo.HexToAddress("0x406f530cc683E668d74F76d13da1Ec5E8cE582ea")

// TODO: move to shared package
func retAddr(ret map[string]any, name string) *ethgo.Address {
	if val, ok := ret[name].(ethgo.Address); ok {
		return &val
	}
	return nil
}

func TestMinting(t *testing.T) {

	chainUrl := "https://bepolia.rpc.berachain.com"
	client, err := jsonrpc.NewClient(chainUrl)
	require.NoError(t, err)

	var skDev2 *ecdsa.PrivateKey
	if skDev2, err = crypto.SKFromHex(os.Getenv("SK_HEX_DEV2")); err != nil {
		return
	}

	var skDev1 *ecdsa.PrivateKey
	if skDev1, err = crypto.SKFromHex(os.Getenv("SK_HEX_DEV1")); err != nil {
		return
	}

	kDev1 := &chain.EcdsaKey{SK: skDev1}
	kDev2 := &chain.EcdsaKey{SK: skDev2}

	mm, err := chain.LoadContract(client, "MasterMinter", kDev1, bepoliaMasterMinter)
	require.NoError(t, err)

	ret, err := mm.Call("owner", ethgo.Latest)
	require.NoError(t, err)
	ownerAddr := retAddr(ret, "0")
	require.Equal(t, kDev2.Address().String(), ownerAddr.String())

	ft, err := chain.LoadContract(client, "FiatTokenV2_2", kDev1, bepoliaFiatTokenProxy)
	require.NoError(t, err)
	_ = ft

	// beraDev1Addr := ethgo.HexToAddress("0x3544FEc2500E38bc0dcb1c14015ACba62774A21d")

	//err = chain.TxnDoWait(mm.Txn("configureController",
	//	beraDev1Addr,
	//	beraDev1Addr))
	//require.NoError(t, err)

	//configAmount := crypto.Millions(1_000_000)
	//err = chain.TxnDoWait(mm.Txn("configureMinter",
	//	configAmount))
	//require.NoError(t, err)

	retName, err := ft.Call("name", ethgo.Latest)
	require.NoError(t, err)
	_ = retName

	retVersion, err := ft.Call("version", ethgo.Latest)
	require.NoError(t, err)
	_ = retVersion

	mintAmount := crypto.Millions(100_000)
	err = chain.TxnDoWait(ft.Txn("mint",
		kDev1.Address(),
		mintAmount))
	require.NoError(t, err)

}
