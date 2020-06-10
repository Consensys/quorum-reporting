package elasticsearch

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/assert"

	"quorumengineering/quorum-report/types"
)

func Test_NewConfigNoCert(t *testing.T) {
	inputConfig := &types.ElasticsearchConfig{
		Addresses: []string{"url1", "url2"},
		CloudID:   "my-cloud-id",
		Username:  "custom-username",
		Password:  "my-secret-password",
		APIKey:    "randomly-generated-api-key",
	}

	expectedOutput := elasticsearch7.Config{
		Addresses: []string{"url1", "url2"},
		CloudID:   "my-cloud-id",
		Username:  "custom-username",
		Password:  "my-secret-password",
		APIKey:    "randomly-generated-api-key",
	}

	outputConfig, err := NewConfig(inputConfig)

	assert.Nil(t, err)
	assert.EqualValues(t, expectedOutput, outputConfig)
}

func Test_NewConfigWithCert(t *testing.T) {
	content := []byte("temporary file's content")
	tmpfile, err := ioutil.TempFile("", "example")
	assert.Nil(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	assert.Nil(t, err)
	err = tmpfile.Close()
	assert.Nil(t, err)

	inputConfig := &types.ElasticsearchConfig{
		Addresses: []string{"url1", "url2"},
		CloudID:   "my-cloud-id",
		Username:  "custom-username",
		Password:  "my-secret-password",
		APIKey:    "randomly-generated-api-key",
		CACert:    tmpfile.Name(),
	}

	expectedOutput := elasticsearch7.Config{
		Addresses: []string{"url1", "url2"},
		CloudID:   "my-cloud-id",
		Username:  "custom-username",
		Password:  "my-secret-password",
		APIKey:    "randomly-generated-api-key",
		CACert:    content,
	}

	outputConfig, err := NewConfig(inputConfig)

	assert.Nil(t, err)
	assert.EqualValues(t, expectedOutput, outputConfig)
}

func Test_NewConfigWithCertReadError(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "example")
	assert.Nil(t, err)
	err = tmpfile.Close()
	assert.Nil(t, err)
	err = os.Remove(tmpfile.Name())
	assert.Nil(t, err)

	inputConfig := &types.ElasticsearchConfig{
		Addresses: []string{"url1", "url2"},
		CloudID:   "my-cloud-id",
		Username:  "custom-username",
		Password:  "my-secret-password",
		APIKey:    "randomly-generated-api-key",
		CACert:    tmpfile.Name(),
	}

	_, err = NewConfig(inputConfig)

	assert.EqualError(t, err, fmt.Sprintf("open %s: no such file or directory", tmpfile.Name()))
}
