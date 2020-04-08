package types

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/naoina/toml"
	"github.com/stretchr/testify/assert"
)

func TestParsePermissionConfig(t *testing.T) {
	d, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(d)

	_, err := ReadConfig(d)
	assert.True(t, err != nil, "expected file not there error")

	fileName := d + "/config.toml"
	_, err = os.Create(fileName)
	_, err = ReadConfig(d)
	assert.True(t, err != nil, "expected unmarshalling error")

	var tmpConfigData ReportInputStruct

	tmpConfigData.Title = "Quorum reporting confg example"
	tmpConfigData.Reporting.RPCAddr = "ws://localhost:23000"
	tmpConfigData.Reporting.GraphQLUrl = "http://localhost:8547/graphql"
	tmpConfigData.Reporting.RPCAddr = "localhost:6666"
	tmpConfigData.Reporting.RPCCorsList = append(tmpConfigData.Reporting.RPCCorsList, "localhost")
	tmpConfigData.Reporting.RPCVHosts = append(tmpConfigData.Reporting.RPCVHosts, "localhost")

	blob, err := toml.Marshal(tmpConfigData)
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("Error writing new node info to file", "fileName", fileName, "err", err)
	}
	_, err = ReadConfig(fileName)
	assert.True(t, err == nil, "error reading the file")
	tmpConfigData.Reporting.MaxReconnectTries = 5
	blob, err = toml.Marshal(tmpConfigData)
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("Error writing new node info to file", "fileName", fileName, "err", err)
	}
	_, err = ReadConfig(fileName)
	assert.True(t, err.Error() == "reconnection details not set properly in the config file", "expected error not thrown")

	tmpConfigData.Reporting.ReconnectInterval = 10
	blob, err = toml.Marshal(tmpConfigData)
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("Error writing new node info to file", "fileName", fileName, "err", err)
	}
	_, err = ReadConfig(fileName)
	assert.True(t, err == nil, "errors encountered")
}
