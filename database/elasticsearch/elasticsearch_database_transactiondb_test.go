package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearch_mocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

//Tests

var (
	//testTransaction = types.Block{
	//	Hash:        common.HexToHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
	//	ParentHash:  common.HexToHash("0x5cde17410e3bb729f745870e166a767bcf07287c0f80bbcb38303eba8dbe5053"),
	//	StateRoot:   common.HexToHash("0x309e12409dc1ff594e12ed7baf41a9190385bc7e32f9c0926dccd95f0a8f62f6"),
	//	TxRoot:      common.HexToHash("0x6473d4f7a3a5638e56fec88a60f765ed321bd15dfade365f232d0b1250a42de0"),
	//	ReceiptRoot: common.HexToHash("0xe65c3585a018f660d1457358967875d1526ebab3e1ce8198757585217fc013b8"),
	//	Number:      10,
	//	GasLimit:    50,
	//	GasUsed:     50,
	//	Timestamp:   100,
	//	ExtraData:   hexutil.Bytes(common.Hex2Bytes("extradata")),
	//	Transactions: []common.Hash{
	//		common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
	//		common.HexToHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
	//		common.HexToHash("0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0"),
	//	},
	//}
	testTransaction = types.Transaction{
		Hash:              common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
		Status:            true,
		BlockNumber:       1,
		Index:             0,
		Nonce:             4,
		From:              common.HexToAddress("0x586e8164bc8863013fe8f1b82092b028a5f8afad"),
		To:                common.HexToAddress("0xcc11df45aba0a4ff198b18300d0b148ad2468834"),
		Value:             10,
		Gas:               30,
		GasUsed:           20,
		CumulativeGasUsed: 40,
		CreatedContract:   common.HexToAddress("0x67bb49f7bd40b6a1226d77dc07fb38f03680c94f"),
		Data:              common.Hex2Bytes("0x4ae157f8a703379222a96b5c01ec83b11b0a0a579b4abc68a10a4c0e7d"),
		PrivateData:       common.Hex2Bytes("0x6060604052341561000f57600080fd5b60405160208061014983398101"),
		IsPrivate:         true,
		Events:            nil,
		InternalCalls:     nil,
	}
)

func TestElasticsearchDB_WriteTransaction_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	var convertedTx Transaction
	convertedTx.From(&testTransaction)

	req := esapi.IndexRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
		Body:       esutil.NewJSONReader(convertedTx),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.WriteTransaction(&testTransaction)

	assert.EqualError(t, err, "test error", "unexpected error message")
}

func TestElasticsearchDB_WriteTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var convertedTx Transaction
	convertedTx.From(&testTransaction)

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
		Body:       esutil.NewJSONReader(convertedTx),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)

	db, _ := New(mockedClient)

	err := db.WriteTransaction(&testTransaction)

	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_ReadTransaction_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(req)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, tx, "unexpected transaction return value")
	assert.EqualError(t, err, "test error")
}

func TestElasticsearchDB_ReadTransaction_WithErrorUnmarshalling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(req)).Return([]byte("{invalid"), nil)

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, tx, "unexpected transaction return value")
	assert.EqualError(t, err, "invalid character 'i' looking for beginning of object key string")
}

func TestElasticsearchDB_ReadTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var convertedTx Transaction
	convertedTx.From(&testTransaction)
	asJson, _ := json.Marshal(convertedTx)

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.GetRequest{
		Index:      TransactionIndex,
		DocumentID: testTransaction.Hash.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(req)).
		Return([]byte(fmt.Sprintf(`{"_source": %s}`, asJson)), nil)

	db, _ := New(mockedClient)

	tx, err := db.ReadTransaction(testTransaction.Hash)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, tx, &testTransaction, "unexpected transaction returned")
}
