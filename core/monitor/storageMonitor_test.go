package monitor

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/state"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

func TestPullStorage(t *testing.T) {
	mockRPC := map[string]interface{}{
		"debug_dumpBlock0x29a": &state.Dump{
			Root: "publicStateRoot",
		},
		"debug_dumpBlock0x29aprivate": &state.Dump{
			Root: "privateStateRoot",
		},
	}
	block := &types.Block{
		Number: 666,
	}
	sm := NewStorageMonitor(nil, client.NewStubQuorumClient(nil, nil, mockRPC))
	err := sm.PullStorage(block)
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
	if block.PublicState.Root != "publicStateRoot" {
		t.Fatalf("expected block public state root as %v, but got %v", "publicStateRoot", block.PublicState.Root)
	}
	if block.PrivateState.Root != "privateStateRoot" {
		t.Fatalf("expected block private state root as %v, but got %v", "privateStateRoot", block.PrivateState.Root)
	}
}
