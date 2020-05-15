package parsers

import "math/big"

func ParseInt(bytes []byte) *big.Int {
	isPositive := bytes[0] < 128

	if isPositive {
		i := new(big.Int)
		i.SetBytes(bytes)
		return i
	}

	//
	for i := 0; i < len(bytes); i++ {
		bytes[i] = ^bytes[i]
	}

	i := new(big.Int)
	i.SetBytes(bytes)
	i.Add(i, new(big.Int).SetUint64(1))
	i.Neg(i)

	return i
}

func ParseUint(bytes []byte) *big.Int {
	return new(big.Int).SetBytes(bytes)
}
