package rpc

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestParseInt_Positive(t *testing.T) {
	storageValue := "0000000000000000000000000000000000000000000000000000000000000041"
	namedType := SolidityTypeEntry{}

	result := parseInt(storageValue, namedType)

	expectedResult := new(big.Int).SetInt64(65)

	assert.Equal(t, expectedResult, result)
}

func TestParseInt_Negative(t *testing.T) {
	storageValue := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd6"
	namedType := SolidityTypeEntry{}

	result := parseInt(storageValue, namedType)

	expectedResult := new(big.Int).SetInt64(-42)

	assert.Equal(t, expectedResult, result)
}

func TestParseInt_Zero(t *testing.T) {
	storageValue := "0000000000000000000000000000000000000000000000000000000000000000"
	namedType := SolidityTypeEntry{}

	result := parseInt(storageValue, namedType)

	expectedResult := new(big.Int)

	assert.Equal(t, expectedResult, result)
}
