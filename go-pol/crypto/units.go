package crypto

import "math/big"

func convert(val uint64, decimals int64) *big.Int {
	v := big.NewInt(int64(val))
	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil)
	return v.Mul(v, exp)
}

func Millions(i uint64) *big.Int {
	return convert(i, 6)
}
