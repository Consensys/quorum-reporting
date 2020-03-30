package rpc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
)

const validABI = `
[
	{ "type" : "function", "name" : "balance", "constant" : true },
	{ "type" : "function", "name" : "send", "constant" : false, "inputs" : [ { "name" : "amount", "type" : "uint256" } ] }
]`

func TestAPIValidation(t *testing.T) {
	var err error
	db := database.NewMemoryDB()
	apis := NewRPCAPIs(db)
	// Test AddAddress validation
	err = apis.AddAddress(common.Address{0})
	if err.Error() != "invalid input" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
	// Test AddContractABI string to ABI parsing
	err = apis.AddContractABI(common.Address{1}, validABI)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	err = apis.AddContractABI(common.Address{1}, "hello")
	if err.Error() != "invalid character 'h' looking for beginning of value" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
}
