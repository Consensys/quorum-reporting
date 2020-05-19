package parsers

import "math/big"

func (p *Parser) ParseInt(bytes []byte) *big.Int {
	//2s complement, so negative is Most Significant Bit is set
	isPositive := bytes[0] < 128

	if isPositive {
		i := new(big.Int)
		i.SetBytes(bytes)
		return i
	}

	// negative, so invert all the bits, add 1 and flip the sign
	for i := 0; i < len(bytes); i++ {
		bytes[i] = ^bytes[i]
	}

	i := new(big.Int)
	i.SetBytes(bytes)
	i.Add(i, new(big.Int).SetUint64(1))
	i.Neg(i)

	return i
}

func (p *Parser) ParseUint(bytes []byte) *big.Int {
	return new(big.Int).SetBytes(bytes)
}
