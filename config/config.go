package config

// ProxyConfig ...
type ProxyConfig struct {
	ProxyName   string     `yaml:"proxy_name"`
	ProxyAddres string     `yaml:"proxy_address"`
	IdleTimeout int        `yaml:"idle_timeout"`
	TargetHost  string     `yaml:"target_host"`
	TargetPort  string     `yaml:"target_port"`
	PrintLogs   bool       `yaml:"print_logs"`
	CompactLogs bool       `yaml:"compact_logs"`
	DNSConfig   *DNSConfig `yaml:"dns_config"`
}

// DNSConfig ...
type DNSConfig struct {
	RefreshRate int    `yaml:"refresh_rate"`
	NeedRefresh bool   `yaml:"need_refresh"`
	BalancerAlg string `yaml:"balancer_alg"` // none, random, round_robin (default none)
}

// SetDefaults sets default values
func (c *ProxyConfig) SetDefaults() {
	if c.ProxyName == "" {
		c.ProxyName = "h2-proxy"
	}

	if c.ProxyAddres == "" {
		c.ProxyAddres = "0.0.0.0:50060"
	}

	if c.IdleTimeout == 0 {
		// value in minutes
		c.IdleTimeout = 5
	}

	if c.DNSConfig == nil {
		c.DNSConfig = &DNSConfig{
			// value in seconds
			RefreshRate: 60,
			NeedRefresh: true,
			BalancerAlg: "round_robin",
		}
	}
}
