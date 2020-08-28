package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"quorumengineering/quorum-report/config"
	"strings"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"

	"quorumengineering/quorum-report/database"
)

//go:generate mockgen -destination=./mocks/api_client_mock.go -package elasticsearch_mocks . APIClient
//go:generate mockgen -destination=./mocks/bulkindexer_mock.go -package elasticsearch_mocks github.com/elastic/go-elasticsearch/v7/esutil BulkIndexer
type APIClient interface {
	ScrollAllResults(index string, query string) ([]interface{}, error)
	// DoRequest executes any operation type for ElasticSearch
	DoRequest(req esapi.Request) ([]byte, error)
	GetBulkHandler(index string) esutil.BulkIndexer
	// CloseIndexers close all bulk update indexers
	CloseIndexers()
}

type DefaultAPIClient struct {
	client   *elasticsearch7.Client
	indexers map[string]esutil.BulkIndexer
}

func NewAPIClient(client *elasticsearch7.Client) (*DefaultAPIClient, error) {
	apiClient := &DefaultAPIClient{
		client:   client,
		indexers: make(map[string]esutil.BulkIndexer),
	}

	for _, idx := range AllIndexes {
		indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Index:         idx,         // The default index name
			Client:        client,      // The Elasticsearch client
			NumWorkers:    1,           // The number of worker goroutines
			FlushBytes:    1024 * 1024, // The flush threshold in bytes
			FlushInterval: time.Second,
		})
		if err != nil {
			return nil, err
		}

		apiClient.indexers[idx] = indexer
	}

	return apiClient, nil
}

func NewClient(config elasticsearch7.Config) (*elasticsearch7.Client, error) {
	return elasticsearch7.NewClient(config)
}

func NewConfig(config *config.ElasticsearchConfig) (elasticsearch7.Config, error) {
	var cert []byte
	if config.CACert != "" {
		certificate, err := ioutil.ReadFile(config.CACert)
		if err != nil {
			return elasticsearch7.Config{}, err
		}
		cert = certificate
	}

	return elasticsearch7.Config{
		Addresses: config.Addresses,
		CloudID:   config.CloudID,

		Username: config.Username,
		Password: config.Password,
		APIKey:   config.APIKey,

		CACert: cert,
	}, nil
}

func (c *DefaultAPIClient) ScrollAllResults(index string, query string) ([]interface{}, error) {
	var (
		scrollID string
		results  []interface{}
	)

	res, _ := c.client.Search(
		c.client.Search.WithIndex(index),
		c.client.Search.WithSort("_doc"),
		c.client.Search.WithSize(10),
		c.client.Search.WithScroll(time.Minute),
		c.client.Search.WithBody(strings.NewReader(query)),
	)

	// Handle the first batch of data and extract the scrollID
	//
	var response map[string]interface{}
	_ = json.NewDecoder(res.Body).Decode(&response)
	res.Body.Close()

	scrollID = response["_scroll_id"].(string)
	hits := response["hits"].(map[string]interface{})["hits"].([]interface{})
	results = append(results, hits...)

	// Perform the scroll requests in sequence
	//
	for {
		// Perform the scroll request and pass the scrollID and scroll duration
		res, err := c.client.Scroll(
			c.client.Scroll.WithScrollID(scrollID),
			c.client.Scroll.WithScroll(time.Minute),
		)
		if err != nil {
			return nil, err
		}
		if res.IsError() {
			err := c.extractError(res.StatusCode, res.Body)
			res.Body.Close()
			return nil, err
		}

		var scrollResponse map[string]interface{}
		_ = json.NewDecoder(res.Body).Decode(&scrollResponse)
		res.Body.Close()

		// Extract the scrollID from response
		scrollID = scrollResponse["_scroll_id"].(string)

		// Extract the search results
		hits := scrollResponse["hits"].(map[string]interface{})["hits"].([]interface{})

		// Break out of the loop when there are no results
		if len(hits) == 0 {
			break
		}

		results = append(results, hits...)
	}

	return results, nil
}

func (c *DefaultAPIClient) DoRequest(req esapi.Request) ([]byte, error) {
	res, err := req.Do(context.TODO(), c.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, c.extractError(res.StatusCode, res.Body)
	}
	return ioutil.ReadAll(res.Body)
}

func (c *DefaultAPIClient) GetBulkHandler(index string) esutil.BulkIndexer {
	return c.indexers[index]
}

func (c *DefaultAPIClient) CloseIndexers() {
	for _, index := range c.indexers {
		index.Close(context.Background())
	}
}

func (c *DefaultAPIClient) extractError(statusCode int, body io.ReadCloser) error {
	var raw map[string]interface{}
	err := json.NewDecoder(body).Decode(&raw)
	if err != nil {
		return ErrCouldNotResolveResp
	}

	// An error occurred with the request
	if raw["error"] != nil {
		errorObj, ok := raw["error"].(map[string]interface{})
		if ok {
			if errorObj["type"] == "index_not_found_exception" {
				return ErrIndexNotFound
			}
			errorStr := fmt.Sprintf("[%d] %s: %s", statusCode, errorObj["type"], errorObj["reason"])
			return fmt.Errorf("error response from Elasticsearch: %s", errorStr)
		}
		// It is possible that the error is just a string not a map
		errorStr, ok := raw["error"].(string)
		if ok {
			return fmt.Errorf("error response from Elasticsearch: %s", errorStr)
		}
	}
	// This was a search request that had no result
	return database.ErrNotFound
}
