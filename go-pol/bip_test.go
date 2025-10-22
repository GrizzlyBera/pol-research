package go_pol

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/wallet"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestBase12Words(t *testing.T) {

	testEntropy, _ := bip39.NewEntropy(128)
	testSeedPhrase, err := bip39.NewMnemonic(testEntropy)
	require.NoError(t, err)

	checkEntropy, err := bip39.EntropyFromMnemonic(testSeedPhrase)
	require.NoError(t, err)

	require.Equal(t, testEntropy, checkEntropy)

	addrFromSeedPhrase := func(seedPhrase string) ethgo.Address {
		seed := bip39.NewSeed(seedPhrase, "")
		masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
		require.NoError(t, err)
		priv, err := wallet.DefaultDerivationPath.Derive(masterKey)
		require.NoError(t, err)
		return wallet.NewKey(priv).Address()
	}
	testAddr := addrFromSeedPhrase(testSeedPhrase)

	seedWords := strings.Split(testSeedPhrase, " ")

	start := time.Now()

	cntTrys := 0
	for {
		for i := range seedWords {
			j := rand.Intn(i + 1)
			seedWords[i], seedWords[j] = seedWords[j], seedWords[i]
		}
		checkAddr := addrFromSeedPhrase(strings.Join(seedWords, " "))
		if testAddr == checkAddr {
			break
		}
		cntTrys++
		if cntTrys%1_000 == 0 {
			println(cntTrys / 1_000)
			if cntTrys == 10_000 {
				break
			}
		}
	}

	elapsed := time.Since(start)
	// 13_479_765_292
	// 12! is 479_001_600
	println(elapsed)

	// rubber eagle card blossom library alter merry fetch before whisper spider husband
	println(testSeedPhrase)

}
