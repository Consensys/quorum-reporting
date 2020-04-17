package elasticsearch

import (
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"

	"quorumengineering/quorum-report/types"
)

func NewClient(config elasticsearch7.Config) *elasticsearch7.Client {
	//TODO: handle error
	client, _ := elasticsearch7.NewClient(config)
	return client
}

func NewConfig(config types.ElasticSearchConfig) elasticsearch7.Config {
	return elasticsearch7.Config{
		Addresses: config.Addresses,
		CloudID:   config.CloudID,
	}
}

//////
//
//client := elasticsearch.NewClient()
////info, _ := client.Info()
////i, _ := ioutil.ReadAll(info.Body)
////fmt.Println(string(i))
////info.Body.Close()
//
////	{
////		"query": { "match_all": {} },
////		"sort": [
////			{ "account_number": "asc" }
////		]
////	}
//
//var buf bytes.Buffer
//query := map[string]interface{}{
//"query": map[string]interface{}{
//"match_all": map[string]interface{}{
//},
//},
//"sort": map[string]interface{}{
//"account_number": "asc",
//},
//}
//if err := json.NewEncoder(&buf).Encode(query); err != nil {
//log.Fatalf("Error encoding query: %s", err)
//}
//
//// Perform the search request.
//res, err := client.Search(
//client.Search.WithContext(context.Background()),
//client.Search.WithIndex("bank"),
//client.Search.WithBody(&buf),
//client.Search.WithTrackTotalHits(true),
//client.Search.WithPretty(),
//)
//if err != nil {
//log.Fatalf("Error getting response: %s", err)
//}
//defer res.Body.Close()
//
//i, _ := ioutil.ReadAll(res.Body)
//fmt.Println(string(i))
//
////esapi.SearchRequest{}
////client.Search()
///////
