package elasticsearch

import (
	"bytes"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
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
	req  esapi.SearchRequest
	body string
}

func NewSearchRequestMatcher(req esapi.SearchRequest) *SearchRequestMatcher {
	body, _ := ioutil.ReadAll(req.Body)
	return &SearchRequestMatcher{req: req, body: string(body)}
}

func (rm *SearchRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.SearchRequest); ok {
		actualBody, _ := ioutil.ReadAll(val.Body)
		//only ever expect one item here, and to always be populated
		return val.Index[0] == rm.req.Index[0] &&
			*val.From == *rm.req.From &&
			*val.Size == *rm.req.Size &&
			// check contents of "sort" field
			string(actualBody) == rm.body
	}
	return false
}

func (rm *SearchRequestMatcher) String() string {
	return fmt.Sprintf("SearchRequestMatcher{%s/%d/%d/%s/%s}", rm.req.Index, rm.req.From, rm.req.Size, rm.req.Sort, rm.body)
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

type GetRequestMatcher struct {
	req esapi.GetRequest
}

func NewGetRequestMatcher(req esapi.GetRequest) *GetRequestMatcher {
	return &GetRequestMatcher{req: req}
}

func (rm *GetRequestMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esapi.GetRequest); ok {
		return val.Index == rm.req.Index && val.DocumentID == rm.req.DocumentID
	}
	return false
}

func (rm *GetRequestMatcher) String() string {
	return fmt.Sprintf("GetRequestMatcher{%s/%s}", rm.req.Index, rm.req.DocumentID)
}

type BulkIndexItemMatcher struct {
	item esutil.BulkIndexerItem
	body string
}

func NewBulkIndexerItemMatcher(item esutil.BulkIndexerItem) *BulkIndexItemMatcher {
	data, _ := ioutil.ReadAll(item.Body)
	return &BulkIndexItemMatcher{
		item: item,
		body: string(data),
	}
}

func (bim *BulkIndexItemMatcher) Matches(x interface{}) bool {
	if val, ok := x.(esutil.BulkIndexerItem); ok {
		valBody, _ := ioutil.ReadAll(val.Body)
		return val.Index == bim.item.Index &&
			val.DocumentID == bim.item.DocumentID &&
			string(valBody) == bim.body
	}
	return false
}

func (bim *BulkIndexItemMatcher) String() string {
	return fmt.Sprintf("BulkIndexItemMatcher{%s/%s/%s}", bim.item.Index, bim.item.DocumentID, bim.body)
}
