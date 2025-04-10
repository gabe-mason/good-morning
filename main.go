package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gabe-mason/good-morning/agent"
	"github.com/gabe-mason/good-morning/config"
	"github.com/gabe-mason/good-morning/tools"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicAPIKey))

	agent := agent.NewAgent(client, tools.ToolCalls{
		tools.NewCalendar(cfg.ICSURL),
		tools.NewGithub(cfg.GithubToken),
		tools.NewLinear(cfg.LinearToken),
	}, cfg)

	summary, err := agent.GenerateDailySummary(ctx)
	if err != nil {
		panic(err)
	}

	// Write summary to file
	if err := os.WriteFile(cfg.GetSummaryLocation(), []byte(summary), 0644); err != nil {
		panic(fmt.Errorf("failed to write summary: %v", err))
	}
}
