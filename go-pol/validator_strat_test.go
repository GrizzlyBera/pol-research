package go_pol

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/GrizzlyBera/pol-research/go-pol/chain"
	"github.com/GrizzlyBera/pol-research/go-pol/crypto"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/jsonrpc"
	"math/big"
	"os"
	"testing"
)

func retBigInt(ret map[string]any, name string) *big.Int {
	if val, ok := ret[name].(*big.Int); ok {
		return val
	}
	return nil
}

func retAddr(ret map[string]any, name string) *ethgo.Address {
	if val, ok := ret[name].(ethgo.Address); ok {
		return &val
	}
	return nil
}

func retBool(ret map[string]any, name string) *bool {
	if val, ok := ret[name].(bool); ok {
		return &val
	}
	return nil
}

func TestBasicValidatorStrat(t *testing.T) {

	chainUrl := "http://34.47.11.184:8545"
	client, err := jsonrpc.NewClient(chainUrl)
	require.NoError(t, err)

	var sk *ecdsa.PrivateKey
	sk, err = crypto.SKFromHex("00")
	require.NoError(t, err)

	k := &chain.EcdsaKey{SK: sk}

	rvf, err := chain.LoadContract(client, "RewardVaultFactory", k, chain.RewardVaultFactoryAddr)
	require.NoError(t, err)

	bc, err := chain.LoadContract(client, "BeraChef", k, chain.BeraChefAddr)
	require.NoError(t, err)

	ret, err := rvf.Call("allVaultsLength", ethgo.Latest)
	require.NoError(t, err)

	cntVaults := retBigInt(ret, "0")
	require.NotNil(t, cntVaults)

	whiteListedVaults := make([]ethgo.Address, 0)

	for i := int64(0); i < cntVaults.Int64(); i++ {
		// check whitelist
		ret, err = rvf.Call("allVaults", ethgo.Latest, big.NewInt(i))
		require.NoError(t, err)
		vaultAddr := retAddr(ret, "0")
		require.NotNil(t, vaultAddr)

		ret, err = bc.Call("isWhitelistedVault", ethgo.Latest, *vaultAddr)
		require.NoError(t, err)
		isWhitelistedVault := retBool(ret, "0")
		require.NotNil(t, isWhitelistedVault)
		if *isWhitelistedVault {
			whiteListedVaults = append(whiteListedVaults, *vaultAddr)
		}

		// TESTING
		if i%50 == 0 {
			println(fmt.Sprintf("%d vaults out of %d", i, cntVaults.Int64()))
		}
	}

	println(len(whiteListedVaults))

	// add known vault for testing...
	whiteListedVaults = append(whiteListedVaults, ethgo.HexToAddress("0x6679F737923f0F99A50F7A1a3D4f8092BE11795C"))

	tempDir, err := os.MkdirTemp("", "test-temp-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir) // Ensure the temporary directory is removed when done

	rv, err := chain.LoadContract(client, "RewardVault", k, whiteListedVaults[0])
	require.NoError(t, err)

	testEvt := rv.GetABI().Events["BGTBoosterIncentivesProcessed"]
	testEvts := []*abi.Event{testEvt}

	currentBlock, err := client.Eth().BlockNumber()
	require.NoError(t, err)

	handleAdded := func(log *ethgo.Log) error {
		return nil
	}
	for idx := range whiteListedVaults {
		println(whiteListedVaults[idx].String())
		w := chain.NewWatcher(whiteListedVaults[idx], testEvts, chainUrl, tempDir)
		w.Start(currentBlock, handleAdded)
	}

	select {}
}
