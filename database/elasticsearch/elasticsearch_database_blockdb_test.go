package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearch_mocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

//Tests

var (
	testBlock = types.Block{
		Hash:        common.HexToHash("0x4b603921305ebaa48d863b9f577059a63c653cd8e952372622923708fb657806"),
		ParentHash:  common.HexToHash("0x5cde17410e3bb729f745870e166a767bcf07287c0f80bbcb38303eba8dbe5053"),
		StateRoot:   common.HexToHash("0x309e12409dc1ff594e12ed7baf41a9190385bc7e32f9c0926dccd95f0a8f62f6"),
		TxRoot:      common.HexToHash("0x6473d4f7a3a5638e56fec88a60f765ed321bd15dfade365f232d0b1250a42de0"),
		ReceiptRoot: common.HexToHash("0xe65c3585a018f660d1457358967875d1526ebab3e1ce8198757585217fc013b8"),
		Number:      10,
		GasLimit:    50,
		GasUsed:     50,
		Timestamp:   100,
		ExtraData:   hexutil.Bytes(common.Hex2Bytes("extradata")),
		Transactions: []common.Hash{
			common.HexToHash("0xf4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"),
			common.HexToHash("0x693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"),
			common.HexToHash("0x5c83fa5955aff33c61813105851777bcd2adc85deb9af6286ba42c05cd768de0"),
		},
	}
)

func TestElasticsearchDB_WriteBlock_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.EqualError(t, err, "test error", "unexpected error message")
}

func TestElasticsearchDB_WriteBlock_ErrorFetchingLastPersisted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return(nil, errors.New("test error - last persisted"))

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.EqualError(t, err, "test error - last persisted", "unexpected error message")
}

// This test is for where the next block written is not the next after lastPersisted
// i.e. where we left a gap of blocks since the app was shutdown and this block is
// from the current chain head
func TestElasticsearchDB_WriteBlock_NotSequential(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return([]byte(`{"_source": {"lastPersisted": 1}}`), nil)

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.Nil(t, err, "unexpected error")
}

// This test is for where the next block written is after lastPersisted
// but hasn't yet caught up to where the chain head was at application start
// i.e. there is still a gap of blocks
func TestElasticsearchDB_WriteBlock_IsSequentialNotCaughtUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}
	lastPersistedIndexRequest := esapi.IndexRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
		Body:       strings.NewReader(`{"lastPersisted": 10}`),
	}
	readBlockReq := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "11",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return([]byte(`{"_source": {"lastPersisted": 9}}`), nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(readBlockReq)).
		Return(nil, errors.New("test error - not found"))
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(lastPersistedIndexRequest))

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.Nil(t, err, "unexpected error")
}

// This test is for where the next block written is after lastPersisted
// and has caught up to where the chain head was at application start
// which means the "lastPersisted" will be the current chain head
func TestElasticsearchDB_WriteBlock_IsSequentialAndCaughtUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}
	lastPersistedIndexRequest := esapi.IndexRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
		Body:       strings.NewReader(`{"lastPersisted": 11}`),
	}
	readBlockReq1 := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "11",
	}
	readBlockReq2 := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "12",
	}

	testBlockAsJson, _ := json.Marshal(testBlock)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return([]byte(`{"_source": {"lastPersisted": 9}}`), nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(readBlockReq1)).
		Return([]byte(fmt.Sprintf(`{"_source": %s}`, testBlockAsJson)), nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(readBlockReq2)).
		Return(nil, errors.New("test error - not found"))
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(lastPersistedIndexRequest))

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.Nil(t, err, "unexpected error")
}

// This test is for where the next block written is not the next after lastPersisted
// i.e. where we left a gap of blocks since the app was shutdown and this block is
// from the current chain head
func TestElasticsearchDB_WriteBlock_ErrorWritingLastPersisted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	err := db.WriteBlock(&testBlock)

	assert.EqualError(t, err, "test error", "unexpected error")
}

func TestElasticsearchDB_ReadBlock_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	blockReadRequest := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "10",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(blockReadRequest)).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	block, err := db.ReadBlock(10)

	assert.Nil(t, block, "unexpected block returned")
	assert.EqualError(t, err, "test error", "unexpected error")
}

func TestElasticsearchDB_ReadBlock_WithErrorReadingResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	blockReadRequest := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "10",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(blockReadRequest)).
		Return([]byte("{invalid json"), nil)

	db, _ := New(mockedClient)

	block, err := db.ReadBlock(10)

	assert.Nil(t, block, "unexpected block returned")
	assert.EqualError(t, err, "invalid character 'i' looking for beginning of object key string", "unexpected error")
}

func TestElasticsearchDB_ReadBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	blockReadRequest := esapi.GetRequest{
		Index:      BlockIndex,
		DocumentID: "10",
	}
	testBlockAsJson, _ := json.Marshal(testBlock)

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(blockReadRequest)).
		Return([]byte(fmt.Sprintf(`{"_source": %s}`, string(testBlockAsJson))), nil)

	db, _ := New(mockedClient)

	block, err := db.ReadBlock(10)

	assert.Nil(t, err, "unexpected block returned")
	assert.Equal(t, &testBlock, block, "unexpected block output")
}

func TestElasticsearchDB_WriteBlocks_NoBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)
	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test

	db, _ := New(mockedClient)

	err := db.WriteBlocks([]*types.Block{})

	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_WriteBlocks_SingleBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)

	req := esapi.IndexRequest{
		Index:      BlockIndex,
		DocumentID: "10",
		Body:       esutil.NewJSONReader(testBlock),
		Refresh:    "true",
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewIndexRequestMatcher(req)).Return(nil, nil)
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return([]byte(`{"_source": {"lastPersisted": 1}}`), nil)

	db, _ := New(mockedClient)

	err := db.WriteBlocks([]*types.Block{&testBlock})

	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_WriteBlocks_MultipleBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearch_mocks.NewMockAPIClient(ctrl)
	mockedBulkIndexer := elasticsearch_mocks.NewMockBulkIndexer(ctrl)

	var (
		dbBlockOne Block
		dbBlockTwo Block
	)
	blockTwo := testBlock
	p := &blockTwo
	p.Number = testBlock.Number + 1
	dbBlockOne.From(&testBlock)
	dbBlockTwo.From(p)

	req := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: "10",
		Body:       esutil.NewJSONReader(dbBlockOne),
	}
	req2 := esutil.BulkIndexerItem{
		Action:     "create",
		DocumentID: "11",
		Body:       esutil.NewJSONReader(dbBlockTwo),
	}
	lastPersistedRequest := esapi.GetRequest{
		Index:      MetaIndex,
		DocumentID: "lastPersisted",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().GetBulkHandler(BlockIndex).Return(mockedBulkIndexer)
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req, esutil.BulkIndexerResponseItem{})
		})
	mockedBulkIndexer.EXPECT().
		Add(gomock.Any(), NewBulkIndexerItemMatcher(req2)).
		Do(func(ctx context.Context, item esutil.BulkIndexerItem) {
			item.OnSuccess(context.Background(), req2, esutil.BulkIndexerResponseItem{})
		})
	mockedClient.EXPECT().
		DoRequest(NewGetRequestMatcher(lastPersistedRequest)).
		Return([]byte(`{"_source": {"lastPersisted": 1}}`), nil)

	db, _ := New(mockedClient)

	err := db.WriteBlocks([]*types.Block{&testBlock, p})

	assert.Nil(t, err, "unexpected error")
}
