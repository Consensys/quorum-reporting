package parsers

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
	"testing"
)

func Test2(t *testing.T) {

	fmt.Println(hash(22).String())

	a := common.Hex2Bytes("d833147d7dc355ba459fc788f669e58cfaf9dc25ddcd0702e87d69c7b5124289")
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(a)
	b := common.BytesToHash(hasher.Sum(nil))

	fmt.Println(b.String())

}
