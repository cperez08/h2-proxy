package proxy

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProxyGRPC(t *testing.T) {
	cfg := NewProxyGRPC("127.0.0.1", "8080")

	assert.Equal(t, cfg.TargetHost, "127.0.0.1")
	assert.Equal(t, cfg.TargetPort, "8080")
}

func TestNewProxyFromFile(t *testing.T) {
	cfg, err := NewProxyFromFile("../config/config.yaml")
	if err != nil {
		t.Log("unexpected error reading yaml file ", err)
		t.FailNow()
	}

	assert.Equal(t, cfg.ProxyName, "h2-proxy")

	os.Setenv("H2_PROXY_TARGET_HOST", "")
	_, err = NewProxyFromFile("../config/noexists.yaml")
	assert.Error(t, err, "configs cannot be loaded via defaults since H2_PROXY_TARGET_HOST ors H2_PROXY_TARGET_PORT are not set")

	os.Setenv("H2_PROXY_TARGET_HOST", "127.0.0.1")
	os.Setenv("H2_PROXY_TARGET_PORT", "8080")
	cfg, err = NewProxyFromFile("../config/noexists.yaml")
	assert.Equal(t, cfg.TargetHost, "127.0.0.1")

	// force malformed yaml
	fileName := "config2.yaml"
	bt := []byte(`proxy_address: '0.0.0.0:8080'
	proxy_name: 'h2-proxy'
		target_host: '127.0.0.1'
	`)
	CreateTmpFile(fileName, bt)
	_, errMarshal := NewProxyFromFile("../config/" + fileName)
	if errMarshal == nil {
		t.Log("expecting error reading file")
		t.FailNow()
	}

	RemoveTmpFile(fileName)

	// force error by no target host or port
	fileName = "config2.yaml"
	bt = []byte(`proxy_address: '0.0.0.0:8080'`)
	CreateTmpFile(fileName, bt)
	_, err = NewProxyFromFile("../config/" + fileName)
	if errMarshal == nil {
		t.Log("expecting error reading host and port")
		t.FailNow()
	}

	RemoveTmpFile(fileName)
}

func CreateTmpFile(name string, content []byte) {
	ioutil.WriteFile("../config/"+name, content, 0644)
}

func RemoveTmpFile(name string) {
	os.Remove("../config/" + name)
}
