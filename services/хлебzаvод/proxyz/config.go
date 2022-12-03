package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

const configFile = "config.json"

type ProxyConfig struct {
	From string
	To   *url.URL
}

func LoadConfigs() ([]ProxyConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg []struct {
		From string
		To   string
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config file: %w", err)
	}

	proxyCfg := make([]ProxyConfig, len(cfg))
	for i, proxy := range cfg {
		url, err := url.Parse(proxy.To)
		if err != nil {
			return nil, fmt.Errorf("parsing 'to' url (%s) for '%s': %w", proxy.To, proxy.From, err)
		}

		proxyCfg[i] = ProxyConfig{
			From: proxy.From,
			To:   url,
		}
	}
	return proxyCfg, nil
}
