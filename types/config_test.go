package types

import (
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
	if _, err := ReadConfig(fileName); err == nil {
		t.Fatal("expect file not there error")
	}

	// test config file with missing fields
	var tmpConfigData ReportInputStruct
	tmpConfigData.Title = "Quorum reporting config example"
	tmpConfigData.Server.RPCAddr = "ws://localhost:23000"
	tmpConfigData.Connection.GraphQLUrl = "http://localhost:8547/graphql"
	tmpConfigData.Connection.WSUrl = "localhost:6666"
	tmpConfigData.Server.RPCCorsList = append(tmpConfigData.Server.RPCCorsList, "localhost")
	tmpConfigData.Server.RPCVHosts = append(tmpConfigData.Server.RPCVHosts, "localhost")

	blob, err := toml.Marshal(tmpConfigData)
	if err != nil {
		t.Fatal("error marshalling test config file", "error", err)
	}
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("error writing new node info to file", "fileName", fileName, "error", err)
	}
	if _, err := ReadConfig(fileName); err != nil {
		t.Fatal("error reading config file", "error", err)
	}
	tmpConfigData.Connection.MaxReconnectTries = 5
	blob, err = toml.Marshal(tmpConfigData)
	if err != nil {
		t.Fatal("error marshalling test config file", "error", err)
	}
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("error writing new node info to file", "fileName", fileName, "error", err)
	}
	if _, err := ReadConfig(fileName); err.Error() != "reconnection details not set properly in the config file" {
		t.Fatalf("expected %v, but got %v", "reconnection details not set properly in the config file", err)
	}

	tmpConfigData.Connection.ReconnectInterval = 10
	blob, err = toml.Marshal(tmpConfigData)
	if err != nil {
		t.Fatal("error marshalling test config file")
	}
	if err := ioutil.WriteFile(fileName, blob, 0644); err != nil {
		t.Fatal("error writing new node info to file", "fileName", fileName, "error", err)
	}
	if _, err := ReadConfig(fileName); err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}

	// test config.sample.toml is valid
	if _, err := ReadConfig("../config.sample.toml"); err != nil {
		t.Fatal("error reading sample config file")
	}

}
