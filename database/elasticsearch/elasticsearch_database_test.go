package elasticsearch

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// GoMock matchers that can be used to check specific request types

type IndexRequestMatcher struct {
	req esapi.IndexRequest
}

func NewIndexRequestMatcher(req esapi.IndexRequest) *IndexRequestMatcher {
	return &IndexRequestMatcher{req: req}
}

func (rm *IndexRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.IndexRequest); ok {
		actualBody, _ := ioutil.ReadAll(val.Body)
		expectedBody, _ := ioutil.ReadAll(rm.req.Body)
		return val.DocumentID == rm.req.DocumentID &&
			val.Index == rm.req.Index &&
			bytes.Compare(actualBody, expectedBody) == 0
	}
	return false
}

func (rm *IndexRequestMatcher) String() string {
	return fmt.Sprintf("IndexRequestMatcher{}")
}

func TestAddSingleAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      "contract",
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any())
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(ex))

	db := New(mockedClient)

	err := db.AddAddresses([]common.Address{addr})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddMultipleAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr1 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	addr2 := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f35")

	contract1 := Contract{
		Address:             addr1,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}
	req1 := esapi.IndexRequest{
		Index:      "contract",
		DocumentID: addr1.String(),
		Body:       esutil.NewJSONReader(contract1),
		Refresh:    "true",
	}
	contract2 := Contract{
		Address:             addr2,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}
	req2 := esapi.IndexRequest{
		Index:      "contract",
		DocumentID: addr2.String(),
		Body:       esutil.NewJSONReader(contract2),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any())
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(req1))
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(req2))

	db := New(mockedClient)

	err := db.AddAddresses([]common.Address{addr1, addr2})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddNoAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	mockedClient.EXPECT().DoRequest(gomock.Any())

	db := New(mockedClient)

	err := db.AddAddresses([]common.Address{})

	assert.Nil(t, err, "expected error to be nil")
}

func TestAddSingleAddressWithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := NewMockAPIClient(ctrl)

	addr := common.HexToAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	contract := Contract{
		Address:             addr,
		ABI:                 "",
		CreationTransaction: common.Hash{},
		LastFiltered:        0,
	}

	ex := esapi.IndexRequest{
		Index:      "contract",
		DocumentID: addr.String(),
		Body:       esutil.NewJSONReader(contract),
		Refresh:    "true",
	}

	mockedClient.EXPECT().DoRequest(gomock.Any())
	mockedClient.EXPECT().IndexRequest(NewIndexRequestMatcher(ex)).Return(nil, errors.New("test error"))

	db := New(mockedClient)

	err := db.AddAddresses([]common.Address{addr})

	assert.Nil(t, err, "expected error to be nil")
}
