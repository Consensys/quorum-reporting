package types

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"

	"github.com/naoina/toml"
)

func TestConfigFile(t *testing.T) {
	d, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(d)
	fileName := d + "/config.toml"

	// test reading a non-exist file
	_, err := ReadConfig(fileName)
	assert.NotNil(t, err, "expect file not there error")

	// test config file with missing fields
	var tmpConfigData ReportingConfig
	tmpConfigData.Title = "Quorum reporting config example"
	tmpConfigData.Server.RPCAddr = "ws://localhost:23000"
	tmpConfigData.Connection.GraphQLUrl = "http://localhost:8547/graphql"
	tmpConfigData.Connection.WSUrl = "localhost:6666"
	tmpConfigData.Server.RPCCorsList = append(tmpConfigData.Server.RPCCorsList, "localhost")
	tmpConfigData.Server.RPCVHosts = append(tmpConfigData.Server.RPCVHosts, "localhost")
	tmpConfigData.Tuning.BlockProcessingQueueSize = 10

	blob, err := toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)

	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Nil(t, err, "error reading config file: %s", err)

	tmpConfigData.Connection.MaxReconnectTries = 5
	blob, err = toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)

	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Nil(t, err, "expected no error, but got %v", err)

	tmpConfigData.Connection.ReconnectInterval = 10
	blob, err = toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)

	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Nil(t, err, "expected no error, but got %v", err)

	// test config.sample.toml is valid
	_, err = ReadConfig("../config.sample.toml")
	assert.Nil(t, err, "error reading sample config file")
}
