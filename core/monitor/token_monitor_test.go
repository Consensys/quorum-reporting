package monitor

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
	"testing"
)

type CustomEIP165StubClient struct {
	*client.StubQuorumClient
	implementedInterface string
}

func (stub *CustomEIP165StubClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if common.Bytes2Hex(msg.Data)[8:16] == "ffffffff" {
		return common.LeftPadBytes([]byte{}, 32), nil
	}
	if common.Bytes2Hex(msg.Data)[8:16] == "01ffc9a7" {
		return common.LeftPadBytes([]byte{1}, 32), nil
	}
	if common.Bytes2Hex(msg.Data)[8:16] == stub.implementedInterface {
		return common.LeftPadBytes([]byte{1}, 32), nil
	}
	return common.LeftPadBytes([]byte{}, 0), nil
}

func TestDefaultTokenMonitor_InspectAddresses_NoAddresses(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": []byte{},
	}
	stubClient := client.NewStubQuorumClient(nil, nil, mockRPC)

	tokenMonitor := NewDefaultTokenMonitor(stubClient)

	res, err := tokenMonitor.InspectAddresses([]common.Address{}, nil)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(res), "wanted empty list, but got %v", res)
}

func TestDefaultTokenMonitor_InspectAddresses_EIP165WithERC20(t *testing.T) {
	mockCallValue := make([]byte, 32)
	mockCallValue[31] = 1
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": mockCallValue,
	}
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, nil, mockRPC),
		"36372b07",
	}

	tx := &types.Transaction{
		Hash:        common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:   common.HexToHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber: 1,
	}

	tokenMonitor := NewDefaultTokenMonitor(stubClient)
	res, err := tokenMonitor.InspectAddresses([]common.Address{{9, 8, 7}}, tx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res[common.Address{9, 8, 7}], "ERC20")
}

func TestDefaultTokenMonitor_InspectAddresses_EIP165WithERC721(t *testing.T) {
	mockCallValue := make([]byte, 32)
	mockCallValue[31] = 1
	mockRPC := map[string]interface{}{
		"eth_call<ethereum.CallMsg Value><*big.Int Value>": mockCallValue,
	}
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, nil, mockRPC),
		"80ac58cd",
	}

	tx := &types.Transaction{
		Hash:        common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:   common.HexToHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber: 1,
	}

	tokenMonitor := NewDefaultTokenMonitor(stubClient)
	res, err := tokenMonitor.InspectAddresses([]common.Address{{9, 8, 7}}, tx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res[common.Address{9, 8, 7}], "ERC721")
}
