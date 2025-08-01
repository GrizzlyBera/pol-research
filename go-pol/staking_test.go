package go_pol

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

func TestStakingBasic(t *testing.T) {

	client, err := jsonrpc.NewClient("https://bepolia.rpc.berachain.com")
	require.NoError(t, err)

	var sk *ecdsa.PrivateKey
	if sk, err = crypto.SKFromHex(os.Getenv("SK_HEX")); err != nil {
		return
	}

	k := &chain.EcdsaKey{SK: sk}

	staking, err := chain.LoadContract(client, "WBERAStakerVault", k, chain.BepoliaOldStakingAddr)
	require.NoError(t, err)

	beraMainAddr := ethgo.HexToAddress("0x160D0E134b78BF083Bd7f03b5d9Fcbcb1c6FF27A")

	ret, err := staking.Call("balanceOf", ethgo.Latest, beraMainAddr)
	require.NoError(t, err)
	_ = ret

	err = chain.TxnDoWait(staking.Txn("completeWithdrawal", true))
	require.NoError(t, err)
}
