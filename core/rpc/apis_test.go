package rpc

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"quorumengineering/quorum-report/database"
)

func TestAPIValidation(t *testing.T) {
	db := database.NewMemoryDB()
	apis := NewRPCAPIs(db)
	// Test AddAddress validation
	err := apis.AddAddress(common.Address{0})
	if err.Error() != "invalid input" {
		t.Fatalf("expected %v, but got %v", "invalid input", err)
	}
}
