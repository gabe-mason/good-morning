package main

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gabe-mason/good-morning/agent"
	"github.com/gabe-mason/good-morning/config"
	"github.com/gabe-mason/good-morning/git"
	"github.com/gabe-mason/good-morning/tools"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Initialize Git manager
	gitManager, err := git.NewGitManager(cfg.GoodMorningRoot)
	if err != nil {
		panic(fmt.Errorf("failed to initialize git manager: %v", err))
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicAPIKey))

	agent := agent.NewAgent(client, tools.ToolCalls{
		tools.NewCalendar(cfg.ICSURL),
		tools.NewGithub(cfg.GithubToken),
		tools.NewLinear(cfg.LinearToken),
	}, cfg)

	// Generate new summary
	summary, err := agent.GenerateDailySummary(ctx)
	if err != nil {
		panic(err)
	}

	// Get the current file path
	summaryPath := cfg.GetSummaryLocation()

	// If there are user changes, we need to merge them
	if lastUserCommit, err := gitManager.GetLastUserCommit(); err == nil && lastUserCommit != "" {
		// Get the user's version of the file
		userContent, err := gitManager.GetFileContentAtCommit(lastUserCommit, summaryPath)
		if err == nil {
			// TODO: Implement smart merging logic here
			// For now, we'll just append the new content
			summary = userContent + "\n\n---\n\n" + summary
		}
	}

	// Write summary to file
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		panic(fmt.Errorf("failed to write summary: %v", err))
	}

	// Commit the changes
	if err := gitManager.CommitChanges("Daily summary update", "good-morning-agent <agent@good-morning>"); err != nil {
		panic(fmt.Errorf("failed to commit changes: %v", err))
	}
}
