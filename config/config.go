package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	AnthropicAPIKey string
	GoodMorningRoot string
	ICSURL          string
	GithubToken     string
	LinearToken     string
	LinearTeams     string
	MyName          string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	cfg.AnthropicAPIKey = os.Getenv("GOOD_MORNING_ANTHROPIC_API_KEY")
	if cfg.AnthropicAPIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	cfg.GoodMorningRoot = os.Getenv("GOOD_MORNING_ROOT")
	if cfg.GoodMorningRoot == "" {
		return nil, fmt.Errorf("GOOD_MORNING_ROOT is not set")
	}
	cfg.ICSURL = os.Getenv("GOOD_MORNING_ICS_URL")
	if cfg.ICSURL == "" {
		return nil, fmt.Errorf("GOOD_MORNING_ICS_URL is not set")
	}
	cfg.GithubToken = os.Getenv("GOOD_MORNING_GITHUB_TOKEN")
	if cfg.GithubToken == "" {
		return nil, fmt.Errorf("GOOD_MORNING_GITHUB_TOKEN is not set")
	}
	cfg.LinearToken = os.Getenv("GOOD_MORNING_LINEAR_TOKEN")
	if cfg.LinearToken == "" {
		return nil, fmt.Errorf("GOOD_MORNING_LINEAR_TOKEN is not set")
	}
	cfg.LinearTeams = os.Getenv("GOOD_MORNING_LINEAR_TEAMS")
	if cfg.LinearTeams == "" {
		return nil, fmt.Errorf("GOOD_MORNING_LINEAR_TEAMS is not set")
	}
	cfg.MyName = os.Getenv("GOOD_MORNING_MY_NAME")
	if cfg.MyName == "" {
		return nil, fmt.Errorf("GOOD_MORNING_MY_NAME is not set")
	}
	return cfg, nil
}

func (cfg *Config) GetContextManagerLocation() string {
	now := time.Now()
	userHome, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(cfg.GoodMorningRoot, fmt.Sprintf("/context/context_manager_%s.json", now.Format("2006-01-02")))
	}
	return filepath.Join(userHome, cfg.GoodMorningRoot, fmt.Sprintf("/context/context_manager_%s.json", now.Format("2006-01-02")))
}

func (cfg *Config) GetSummaryLocation() string {
	now := time.Now()
	userHome, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(cfg.GoodMorningRoot, fmt.Sprintf("%s.md", now.Format("2006-01-02")))
	}
	return filepath.Join(userHome, cfg.GoodMorningRoot, fmt.Sprintf("%s.md", now.Format("2006-01-02")))
}
