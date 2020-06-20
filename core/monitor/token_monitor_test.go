package monitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"quorumengineering/quorum-report/client"
	"testing"
)

func TestDefaultTokenMonitor_InspectAddresses_NoAddresses(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": []byte{},
	}
	stubClient := client.NewStubQuorumClient(nil, nil, mockRPC)

	tokenMonitor := NewDefaultTokenMonitor(stubClient)

	res, err := tokenMonitor.InspectAddresses([]common.Address{}, nil, 10)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(res), "wanted empty list, but got %v", res)
}
