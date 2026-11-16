package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	mu         sync.RWMutex
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Extraction ExtractionConfig `yaml:"extraction"`
	CRM        CRMConfig        `yaml:"crm"`
	Slack      SlackConfig      `yaml:"slack"`
	ICP        ICPConfig        `yaml:"icp"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type ExtractionConfig struct {
	Provider string `yaml:"provider"` // "heuristic", "llm"
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

type CRMConfig struct {
	Mode       string           `yaml:"mode"` // "mock" or "live"
	HubSpot    HubSpotConfig    `yaml:"hubspot"`
	Salesforce SalesforceConfig `yaml:"salesforce"`
	Pipedrive  PipedriveConfig  `yaml:"pipedrive"`
}

type HubSpotConfig struct {
	Enabled     bool   `yaml:"enabled"`
	AccessToken string `yaml:"access_token"`
}

type SalesforceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	InstanceURL string `yaml:"instance_url"`
	AccessToken string `yaml:"access_token"`
}

type PipedriveConfig struct {
	Enabled  bool   `yaml:"enabled"`
	APIToken string `yaml:"api_token"`
}

type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
}

type ICPConfig struct {
	MinQualificationScore int        `yaml:"min_qualification_score" json:"min_qualification_score"`
	Weights               ICPWeights `yaml:"weights" json:"weights"`
}

type ICPWeights struct {
	CorporateDomain int `yaml:"corporate_domain" json:"corporate_domain"`
	SeatVolume      int `yaml:"seat_volume" json:"seat_volume"`
	Seniority       int `yaml:"seniority" json:"seniority"`
	Urgency         int `yaml:"urgency" json:"urgency"`
	Budget          int `yaml:"budget" json:"budget"`
	ProductFit      int `yaml:"product_fit" json:"product_fit"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 9393,
		},
		Database: DatabaseConfig{
			Path: "formpilot.db",
		},
		Extraction: ExtractionConfig{
			Provider: "heuristic",
			Model:    "gpt-4o-mini",
		},
		CRM: CRMConfig{
			Mode: "mock",
			HubSpot: HubSpotConfig{
				Enabled: true,
			},
			Salesforce: SalesforceConfig{
				Enabled: false,
			},
			Pipedrive: PipedriveConfig{
				Enabled: false,
			},
		},
		Slack: SlackConfig{
			Enabled: false,
		},
		ICP: ICPConfig{
			MinQualificationScore: 60,
			Weights: ICPWeights{
				CorporateDomain: 20,
				SeatVolume:      25,
				Seniority:       20,
				Urgency:         15,
				Budget:          10,
				ProductFit:      10,
			},
		},
	}
}

func (c *Config) GetICP() ICPConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ICP
}

func (c *Config) UpdateICP(icp ICPConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ICP = icp
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	expanded := expandEnvVars(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}

	if portStr := os.Getenv("FORMPILOT_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = p
		}
	}
	if host := os.Getenv("FORMPILOT_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if dbPath := os.Getenv("FORMPILOT_DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		cfg.Extraction.APIKey = apiKey
	}
	if hsToken := os.Getenv("HUBSPOT_ACCESS_TOKEN"); hsToken != "" {
		cfg.CRM.HubSpot.AccessToken = hsToken
	}
	if sfToken := os.Getenv("SALESFORCE_ACCESS_TOKEN"); sfToken != "" {
		cfg.CRM.Salesforce.AccessToken = sfToken
	}
	if pdToken := os.Getenv("PIPEDRIVE_API_TOKEN"); pdToken != "" {
		cfg.CRM.Pipedrive.APIToken = pdToken
	}

	return cfg, nil
}

func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([a-zA-Z_0-9]+)(?::-([^}]*))?\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) > 1 {
			varName := sub[1]
			val := os.Getenv(varName)
			if val != "" {
				return val
			}
			if len(sub) > 2 {
				return sub[2]
			}
		}
		return ""
	})
}
