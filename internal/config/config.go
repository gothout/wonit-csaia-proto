package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type CSAConfig struct {
	WebhookID string `json:"webhook_id"`
	Token     string `json:"token"`
}

type ChatvoltConfig struct {
	Token   string `json:"token"`
	AgentID string `json:"agent_id"`
}

type IAConfig struct {
	Chatvolt ChatvoltConfig `json:"chatvolt"`
}

type Config struct {
	Server ServerConfig `json:"server"`
	CSA    CSAConfig    `json:"csa"`
	IA     IAConfig     `json:"ia"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}

	return &cfg, nil
}
