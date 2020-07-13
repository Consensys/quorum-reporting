package monitor

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/types"
)

type CustomEIP165StubClient struct {
	*client.StubQuorumClient
	implementedInterface string
}

func (stub *CustomEIP165StubClient) RPCCall(result interface{}, method string, args ...interface{}) error {
	if method == "eth_call" {
		msg := args[0].(types.EIP165Call)
		if msg.Data[8:16] == "ffffffff" {
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(types.HexData("0000000000000000000000000000000000000000000000000000000000000000")))
			return nil
		}
		if msg.Data[8:16] == "01ffc9a7" {
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(types.HexData("0000000000000000000000000000000000000000000000000000000000000001")))
			return nil
		}
		if string(msg.Data[8:16]) == stub.implementedInterface {
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(types.HexData("0000000000000000000000000000000000000000000000000000000000000001")))
			return nil
		}
	}
	return stub.StubQuorumClient.RPCCall(result, method, args)
}

func TestDefaultTokenMonitor_InspectTransaction_EIP165WithERC20_External(t *testing.T) {
	mockRPC := map[string]interface{}{
		"eth_call<types.EIP165Call Value>0x1": types.HexData("0000000000000000000000000000000000000000000000000000000000000001"),
	}
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, mockRPC),
		"36372b07",
	}

	tx := &types.Transaction{
		Hash:            types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:       types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber:     1,
		CreatedContract: types.NewAddress("987"),
	}

	tokenMonitor := NewDefaultTokenMonitor(stubClient, []TokenRule{{scope: types.AllScope, templateName: "ERC20", eip165: "36372b07"}})
	res, err := tokenMonitor.InspectTransaction(tx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res[types.NewAddress("987")], "ERC20")
}

func TestDefaultTokenMonitor_InspectTransaction_EIP165WithERC20(t *testing.T) {
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, nil),
		"36372b07",
	}

	tx := &types.Transaction{
		Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:   types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber: 1,
		InternalCalls: []*types.InternalCall{
			{
				From: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"),
				To:   types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"),
				Type: "CREATE",
			},
		},
	}

	testMatrix := []struct {
		rule   TokenRule
		result map[types.Address]string
	}{
		{
			TokenRule{scope: types.InternalScope, templateName: "ERC20", eip165: "36372b07"},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		},
		{
			TokenRule{scope: types.AllScope, templateName: "ERC20", eip165: "36372b07"},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		},
		{
			TokenRule{scope: types.ExternalScope, templateName: "ERC20", eip165: "36372b07"},
			map[types.Address]string{},
		},
		{
			TokenRule{scope: types.AllScope, templateName: "ERC20", eip165: "36372b07", deployer: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad")},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		},
		{
			TokenRule{scope: types.InternalScope, templateName: "ERC20", eip165: "36372b07", deployer: types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834")}, //TODO: can this be AllScoped?
			map[types.Address]string{},
		},
	}

	for _, tst := range testMatrix {
		tokenMonitor := NewDefaultTokenMonitor(stubClient, []TokenRule{tst.rule})
		res, err := tokenMonitor.InspectTransaction(tx)

		assert.Nil(t, err)
		assert.Equal(t, len(res), len(tst.result))
		assert.EqualValues(t, tst.result, res)
	}
}

func TestDefaultTokenMonitor_InspectTransaction_EIP165WithERC721_External(t *testing.T) {
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, nil),
		"80ac58cd",
	}

	tx := &types.Transaction{
		Hash:            types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:       types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber:     1,
		CreatedContract: types.NewAddress("987"),
	}

	tokenMonitor := NewDefaultTokenMonitor(stubClient, []TokenRule{{scope: types.AllScope, templateName: "ERC721", eip165: "80ac58cd"}})
	res, err := tokenMonitor.InspectTransaction(tx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res[types.NewAddress("987")], "ERC721")
}

func TestDefaultTokenMonitor_InspectTransaction_EIP165WithERC721(t *testing.T) {
	stubClient := &CustomEIP165StubClient{
		client.NewStubQuorumClient(nil, nil),
		"80ac58cd",
	}

	tx := &types.Transaction{
		Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:   types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber: 1,
		InternalCalls: []*types.InternalCall{
			{
				From: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"),
				To:   types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"),
				Type: "CREATE",
			},
		},
	}

	testMatrix := []struct {
		rule   TokenRule
		result map[types.Address]string
	}{
		{
			TokenRule{scope: types.InternalScope, templateName: "ERC721", eip165: "80ac58cd"},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC721"},
		},
		{
			TokenRule{scope: types.AllScope, templateName: "ERC721", eip165: "80ac58cd"},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC721"},
		},
		{
			TokenRule{scope: types.ExternalScope, templateName: "ERC721", eip165: "80ac58cd"},
			map[types.Address]string{},
		},
		{
			TokenRule{scope: types.AllScope, templateName: "ERC721", eip165: "80ac58cd", deployer: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad")},
			map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC721"},
		},
		{
			TokenRule{scope: types.InternalScope, templateName: "ERC721", eip165: "80ac58cd", deployer: types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834")}, //TODO: can this be AllScoped?
			map[types.Address]string{},
		},
	}

	for _, tst := range testMatrix {
		tokenMonitor := NewDefaultTokenMonitor(stubClient, []TokenRule{tst.rule})
		res, err := tokenMonitor.InspectTransaction(tx)

		assert.Nil(t, err)
		assert.Equal(t, len(res), len(tst.result))
		assert.EqualValues(t, tst.result, res)
	}
}

func TestDefaultTokenMonitor_InspectTransaction_BytecodeInspection(t *testing.T) {
	erc20ContractCode := "6080604052600436106100ba576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806301ffc9a7146100bf57806306fdde0314610123578063095ea7b3146101b357806318160ddd1461021857806323b872dd14610243578063313ce567146102c85780636d4ce63c146102f957806370a082311461036257806395d89b41146103b9578063a9059cbb14610449578063cae9ca51146104ae578063dd62ed3e14610559575b600080fd5b3480156100cb57600080fd5b5061010960048036038101908080357bffffffffffffffffffffffffffffffffffffffffffffffffffffffff191690602001909291905050506105d0565b604051808215151515815260200191505060405180910390f35b34801561012f57600080fd5b50610138610638565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561017857808201518184015260208101905061015d565b50505050905090810190601f1680156101a55780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156101bf57600080fd5b506101fe600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506106d6565b604051808215151515815260200191505060405180910390f35b34801561022457600080fd5b5061022d6107c8565b6040518082815260200191505060405180910390f35b34801561024f57600080fd5b506102ae600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803590602001909291905050506107ce565b604051808215151515815260200191505060405180910390f35b3480156102d457600080fd5b506102dd610a47565b604051808260ff1660ff16815260200191505060405180910390f35b34801561030557600080fd5b5061030e610a5a565b60405180827bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19167bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916815260200191505060405180910390f35b34801561036e57600080fd5b506103a3600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050610b50565b6040518082815260200191505060405180910390f35b3480156103c557600080fd5b506103ce610b98565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561040e5780820151818401526020810190506103f3565b50505050905090810190601f16801561043b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561045557600080fd5b50610494600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190505050610c36565b604051808215151515815260200191505060405180910390f35b3480156104ba57600080fd5b5061053f600480360381019080803573ffffffffffffffffffffffffffffffffffffffff16906020019092919080359060200190929190803590602001908201803590602001908080601f0160208091040260200160405190810160405280939291908181526020018383808284378201915050505050509192919290505050610d9c565b604051808215151515815260200191505060405180910390f35b34801561056557600080fd5b506105ba600480360381019080803573ffffffffffffffffffffffffffffffffffffffff169060200190929190803573ffffffffffffffffffffffffffffffffffffffff169060200190929190505050611039565b6040518082815260200191505060405180910390f35b600060036000837bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19167bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190815260200160002060009054906101000a900460ff169050919050565b60048054600181600116156101000203166002900480601f0160208091040260200160405190810160405280929190818152602001828054600181600116156101000203166002900480156106ce5780601f106106a3576101008083540402835291602001916106ce565b820191906000526020600020905b8154815290600101906020018083116106b157829003601f168201915b505050505081565b600081600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040518082815260200191505060405180910390a36001905092915050565b60025481565b6000816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015801561089a575081600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b80156108a65750600082115b15610a3b57816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540192505081905550816000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000828254039250508190555081600160008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825403925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a360019050610a40565b600090505b9392505050565b600560009054906101000a900460ff1681565b600063dd62ed3e7c01000000000000000000000000000000000000000000000000000000000263095ea7b37c0100000000000000000000000000000000000000000000000000000000026323b872dd7c01000000000000000000000000000000000000000000000000000000000260405180807f7472616e7366657228616464726573732c75696e743235362900000000000000815250601901905060405180910390206370a082317c0100000000000000000000000000000000000000000000000000000000026318160ddd7c0100000000000000000000000000000000000000000000000000000000021818181818905090565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60068054600181600116156101000203166002900480601f016020809104026020016040519081016040528092919081815260200182805460018160011615610100020316600290048015610c2e5780601f10610c0357610100808354040283529160200191610c2e565b820191906000526020600020905b815481529060010190602001808311610c1157829003601f168201915b505050505081565b6000816000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410158015610c865750600082115b15610d9157816000803373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282540392505081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055508273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040518082815260200191505060405180910390a360019050610d96565b600090505b92915050565b600082600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508373ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925856040518082815260200191505060405180910390a38373ffffffffffffffffffffffffffffffffffffffff1660405180807f72656365697665417070726f76616c28616464726573732c75696e743235362c81526020017f616464726573732c627974657329000000000000000000000000000000000000815250602e01905060405180910390207c01000000000000000000000000000000000000000000000000000000009004338530866040518563ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020018481526020018373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001828051906020019080838360005b83811015610fdd578082015181840152602081019050610fc2565b50505050905090810190601f16801561100a5780820380516001836020036101000a031916815260200191505b509450505050506000604051808303816000875af192505050151561102e57600080fd5b600190509392505050565b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050929150505600a165627a7a7230582029a6af78b47fac0041e450a0d3455666f55432a374771f353ff771c79b063e2a0029"
	var erc20Abi, erc721Abi types.ABIStructure
	json.Unmarshal([]byte(`[{"inputs":[{"internalType":"uint256","name":"_value","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"tokenOwner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"tokens","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"tokens","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"tokenOwner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"remaining","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"tokens","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenOwner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"balance","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokens","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokens","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"success","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`), &erc20Abi)
	json.Unmarshal([]byte(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_owner","type":"address"},{"indexed":true,"internalType":"address","name":"_approved","type":"address"},{"indexed":true,"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_owner","type":"address"},{"indexed":true,"internalType":"address","name":"_operator","type":"address"},{"indexed":false,"internalType":"bool","name":"_approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"_from","type":"address"},{"indexed":true,"internalType":"address","name":"_to","type":"address"},{"indexed":true,"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"_approved","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"approve","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"},{"internalType":"address","name":"_operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"_operator","type":"address"},{"internalType":"bool","name":"_approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_from","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"uint256","name":"_tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"stateMutability":"payable","type":"function"}]`), &erc721Abi)

	mockRPC := map[string]interface{}{
		"eth_getCodecc11df45aba0a4ff198b18300d0b148ad24688340xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36": types.HexData(erc20ContractCode),
	}
	stubClient := client.NewStubQuorumClient(nil, mockRPC)

	tx := &types.Transaction{
		Hash:        types.NewHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		BlockHash:   types.NewHash("0xefe5cb8d23d632b5d2cdd9f0a151c4b1a84ccb7afa1c57331009aa922d5e4f36"),
		BlockNumber: 1,
		InternalCalls: []*types.InternalCall{
			{
				From: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"),
				To:   types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"),
				Type: "CREATE",
			},
		},
	}

	testMatrix := []struct {
		rule   TokenRule
		result map[types.Address]string
	}{
		{
			TokenRule{scope: types.InternalScope, templateName: "ERC721", abi: erc721Abi.ToInternalABI()},
			map[types.Address]string{}, //No result as contract is ERC20, not ERC721
		},
		//{
		//	TokenRule{scope: types.InternalScope, templateName: "ERC20", abi: erc20Abi.ToInternalABI()},
		//	map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		//},
		//{
		//	TokenRule{scope: types.AllScope, templateName: "ERC20", abi: erc20Abi.ToInternalABI()},
		//	map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		//},
		//{
		//	TokenRule{scope: types.ExternalScope, templateName: "ERC20", abi: erc20Abi.ToInternalABI()},
		//	map[types.Address]string{},
		//},
		//{
		//	TokenRule{scope: types.AllScope, templateName: "ERC20", deployer: types.NewAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"), abi: erc20Abi.ToInternalABI()},
		//	map[types.Address]string{types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"): "ERC20"},
		//},
		//{
		//	TokenRule{scope: types.InternalScope, templateName: "ERC20", deployer: types.NewAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"), abi: erc20Abi.ToInternalABI()}, //TODO: can this be AllScoped?
		//	map[types.Address]string{},
		//},
	}

	for _, tst := range testMatrix {
		tokenMonitor := NewDefaultTokenMonitor(stubClient, []TokenRule{tst.rule})
		res, err := tokenMonitor.InspectTransaction(tx)

		assert.Nil(t, err)
		assert.Equal(t, len(tst.result), len(res))
		assert.EqualValues(t, tst.result, res)
	}
}
