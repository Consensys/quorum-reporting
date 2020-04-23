package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io"
)

type MockAPIClient struct {
	//ScrollAllResults return values
	results []interface{}

	//DoRequest return values
	body []byte
	err  error
}

func NewMockClient() *MockAPIClient {
	return &MockAPIClient{}
}

func (mc *MockAPIClient) SetScrollAllResultsReturn(results []interface{}) {
	mc.results = results
}

func (mc *MockAPIClient) ScrollAllResults(index string, query io.Reader) []interface{} {
	return mc.results
}

func (mc *MockAPIClient) SetDoRequestReturn(body []byte, err error) {
	mc.body = body
	mc.err = err
}

func (mc *MockAPIClient) DoRequest(req esapi.Request) ([]byte, error) {
	return mc.body, mc.err
}
