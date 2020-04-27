package elasticsearch

import (
	"bytes"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io/ioutil"
)

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

type SearchRequestMatcher struct {
	req esapi.SearchRequest
}

func NewSearchRequestMatcher(req esapi.SearchRequest) *SearchRequestMatcher {
	return &SearchRequestMatcher{req: req}
}

func (rm *SearchRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.SearchRequest); ok {
		actualBody, _ := ioutil.ReadAll(val.Body)
		expectedBody, _ := ioutil.ReadAll(rm.req.Body)
		//only ever expect one item here, and to always be populated
		return val.Index[0] == rm.req.Index[0] &&
			bytes.Compare(actualBody, expectedBody) == 0
	}
	return false
}

func (rm *SearchRequestMatcher) String() string {
	return fmt.Sprintf("SearchRequestMatcher{%s}", rm.req.Index)
}

type DeleteRequestMatcher struct {
	req esapi.DeleteRequest
}

func NewDeleteRequestMatcher(req esapi.DeleteRequest) *DeleteRequestMatcher {
	return &DeleteRequestMatcher{req: req}
}

func (rm *DeleteRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.DeleteRequest); ok {
		return val.Index == rm.req.Index && val.DocumentID == rm.req.DocumentID
	}
	return false
}

func (rm *DeleteRequestMatcher) String() string {
	return fmt.Sprintf("DeleteRequestMatcher{%s}", rm.req.Index)
}

type UpdateRequestMatcher struct {
	req esapi.UpdateRequest
}

func NewUpdateRequestMatcher(req esapi.UpdateRequest) *UpdateRequestMatcher {
	return &UpdateRequestMatcher{req: req}
}

func (rm *UpdateRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.UpdateRequest); ok {
		actualBody, _ := ioutil.ReadAll(val.Body)
		expectedBody, _ := ioutil.ReadAll(rm.req.Body)
		return val.Index == rm.req.Index &&
			val.DocumentID == rm.req.DocumentID &&
			bytes.Compare(actualBody, expectedBody) == 0
	}
	return false
}

func (rm *UpdateRequestMatcher) String() string {
	return fmt.Sprintf("UpdateRequestMatcher{%s}", rm.req.Index)
}
