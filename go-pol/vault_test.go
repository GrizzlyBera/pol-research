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

// deployed mock token: 0xaCdb2757Fc2a8B133031d5200F85EE17602F1Cc2
var mockStakingToken = ethgo.HexToAddress("0xaCdb2757Fc2a8B133031d5200F85EE17602F1Cc2")

var testRewardVault = ethgo.HexToAddress("0xB492BFb0ecB05bb48ecC8528250528D89fdD73c4")

func TestPartialReward(t *testing.T) {
	client, err := jsonrpc.NewClient("https://bepolia.rpc.berachain.com")
	require.NoError(t, err)

	var sk *ecdsa.PrivateKey
	if sk, err = crypto.SKFromHex(os.Getenv("SK_HEX")); err != nil {
		return
	}

	k := &chain.EcdsaKey{SK: sk}

	vault, err := chain.LoadContract(client, "RewardVault", k, ethgo.HexToAddress("0x9c84a17467d0f691b4a6fe6c64fa00edb55d9646"))
	require.NoError(t, err)

	//     function getPartialReward(
	//        address account,
	//        address recipient,
	//        uint256 amount
	//    )
	err = chain.TxnDoWait(vault.Txn("getPartialReward",
		k.Address(),
		ethgo.HexToAddress("0x3544FEc2500E38bc0dcb1c14015ACba62774A21d"),
		ethgo.Ether(1)))
	require.NoError(t, err)

}

func TestDeployVault(t *testing.T) {

	client, err := jsonrpc.NewClient("https://bepolia.rpc.berachain.com")
	require.NoError(t, err)

	var sk *ecdsa.PrivateKey
	if sk, err = crypto.SKFromHex(os.Getenv("SK_HEX")); err != nil {
		return
	}

	k := &chain.EcdsaKey{SK: sk}

	// using Bera Dev 1 - 0x3544FEc2500E38bc0dcb1c14015ACba62774A21d
	//args := []interface{}{"Mock Vault 1", "MXV1"}
	//_, mockAddr, err := chain.DeployContract(client, "MockCoin.sol/MockCoin", k, args)
	//require.NoError(t, err)
	//println(mockAddr.String())

	vaultFactory, err := chain.LoadContract(client, "RewardVaultFactoryAddr", k, chain.RewardVaultFactoryAddr)
	require.NoError(t, err)

	//     function createRewardVault(address stakingToken) external returns (address) {
	err = chain.TxnDoWait(vaultFactory.Txn("createRewardVault", mockStakingToken))
	require.NoError(t, err)

}
