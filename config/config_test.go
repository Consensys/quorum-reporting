package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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
	tmpConfigData.Templates = []*TemplateConfig{
		{
			TemplateName:  "SimpleStorage",
			StorageLayout: "{\"storage\":[{\"astId\":3,\"contract\":\"scripts/simplestorage.sol:SimpleStorage\",\"label\":\"storedData\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"}],\"types\":{\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}",
		},
	}

	blob, err := toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)
	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Error(t, err, "expected error, but got %v", err)

	tmpConfigData.Templates = []*TemplateConfig{
		{
			TemplateName:  "SimpleStorage",
			ABI:           "[{\"constant\":true,\"inputs\":[],\"name\":\"storedData\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_x\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"_initVal\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"valueSet\",\"type\":\"event\"}]",
			StorageLayout: "{\"storage\":[{\"astId\":3,\"contract\":\"scripts/simplestorage.sol:SimpleStorage\",\"label\":\"storedData\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"}],\"types\":{\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}",
		},
	}
	tmpConfigData.Rules = []*RuleConfig{
		{
			Scope:        "invalidScope",
			TemplateName: "SimpleStorage",
		},
	}

	blob, err = toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)
	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Error(t, err, "expected error, but got %v", err)

	tmpConfigData.Rules = []*RuleConfig{
		{
			Scope:        "all",
			TemplateName: "",
		},
	}

	blob, err = toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)
	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	_, err = ReadConfig(fileName)
	assert.Error(t, err, "expected error, but got %v", err)

	tmpConfigData.Rules = []*RuleConfig{
		{
			Scope:        "all",
			TemplateName: "SimpleStorage",
		},
	}

	blob, err = toml.Marshal(tmpConfigData)
	assert.Nil(t, err, "error marshalling test config file: %s", err)
	err = ioutil.WriteFile(fileName, blob, 0644)
	assert.Nil(t, err, "error writing new node info to file %s: %s", fileName, err)

	// test config.sample.toml is valid
	_, err = ReadConfig("../config.sample.toml")
	assert.Nil(t, err, "error reading sample config file")
}
