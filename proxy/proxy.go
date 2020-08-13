package proxy

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cperez08/h2-proxy/config"
	"gopkg.in/yaml.v2"
)

// NewProxyGRPC ...
func NewProxyGRPC(host string, port string) *config.ProxyConfig {
	return &config.ProxyConfig{TargetHost: host, TargetPort: port}
}

// NewProxyFromFile ...
func NewProxyFromFile(file string) (*config.ProxyConfig, error) {
	rs := &config.ProxyConfig{}

	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := loadConfigWithDefaults(rs); err != nil {
				return nil, err
			}

			return rs, nil
		}

		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, rs)
	if err != nil {
		log.Printf("unmarshal: %v", err)
		return nil, err
	}

	if rs.TargetHost == "" || rs.TargetPort == "" {
		return nil, errors.New("target host and target port are mandatory")
	}

	rs.SetDefaults()
	return rs, nil
}

func loadConfigWithDefaults(c *config.ProxyConfig) error {
	host := os.Getenv("H2_PROXY_TARGET_HOST")
	port := os.Getenv("H2_PROXY_TARGET_PORT")
	logs := os.Getenv("H2_PROXY_PRINT_LOGS")

	if strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" {
		return errors.New("configs cannot be loaded via defaults since H2_PROXY_TARGET_HOST ors H2_PROXY_TARGET_PORT are not set")
	}

	c.TargetHost = host
	c.TargetPort = port
	logs = strings.ToLower(logs)

	if logs == "true" || logs == "false" {
		c.PrintLogs, _ = strconv.ParseBool(logs)
	}

	c.SetDefaults()
	return nil
}
