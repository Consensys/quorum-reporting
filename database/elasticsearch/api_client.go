package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"quorumengineering/quorum-report/types"
)

//go:generate mockgen -destination=./api_client_mock_test.go -package elasticsearch . APIClient
type APIClient interface {
	ScrollAllResults(index string, query string) ([]interface{}, error)
	//DoRequest executes any operation type for ElasticSearch
	DoRequest(req esapi.Request) ([]byte, error)
	//IndexRequest specifically executes an ElasticSearch index operation
	IndexRequest(req esapi.IndexRequest) ([]byte, error)
}

type DefaultAPIClient struct {
	client *elasticsearch7.Client
}

func NewAPIClient(client *elasticsearch7.Client) *DefaultAPIClient {
	return &DefaultAPIClient{client: client}
}

func NewClient(config elasticsearch7.Config) (*elasticsearch7.Client, error) {
	return elasticsearch7.NewClient(config)
}

func NewConfig(config *types.ElasticsearchConfig) elasticsearch7.Config {
	return elasticsearch7.Config{
		Addresses: config.Addresses,
		CloudID:   config.CloudID,

		Username: config.Username,
		Password: config.Password,
		APIKey:   config.APIKey,
	}
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
			//log.Println("Finished scrolling")
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

func (c *DefaultAPIClient) IndexRequest(req esapi.IndexRequest) ([]byte, error) {
	return c.DoRequest(req)
}

func (c *DefaultAPIClient) extractError(statusCode int, body io.ReadCloser) error {
	var raw map[string]interface{}
	err := json.NewDecoder(body).Decode(&raw)
	if err != nil {
		return errors.New("could not resolve response body")
	}

	//an error occured with the request
	if raw["error"] != nil {
		errorObj := raw["error"].(map[string]interface{})
		return fmt.Errorf("error: [%d] %s: %s", statusCode, errorObj["type"], errorObj["reason"])
	}
	//this was a search request that had no result
	return errors.New("not found")
}
