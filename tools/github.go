package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/anthropics/anthropic-sdk-go"
)

type Github struct {
	token string
}

func NewGithub(token string) *Github {
	return &Github{
		token: token,
	}
}

type GithubInput struct {
	Action string `json:"action" jsonschema_description:"The action to perform (list_my_prs, list_review_requests)"`
}

func (g *Github) Run(ctx context.Context, arguments json.RawMessage) (string, error) {
	var input GithubInput
	if err := json.Unmarshal(arguments, &input); err != nil {
		return "", &InvalidToolArgumentsError{
			ToolName: g.Name(),
			Message:  "invalid JSON format",
		}
	}

	switch input.Action {
	case "list_my_prs":
		return g.listMyPRs(ctx)
	case "list_review_requests":
		return g.listReviewRequests(ctx)
	default:
		return "", &InvalidToolArgumentsError{
			ToolName: g.Name(),
			Message:  "invalid action, supported actions are: list_my_prs, list_review_requests",
		}
	}
}

func (g *Github) searchGitHub(ctx context.Context, query string) (string, error) {
	escapedQuery := url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.github.com/search/issues?q=%s", escapedQuery), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	jsonResponse, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}

	return string(jsonResponse), nil
}

func (g *Github) listMyPRs(ctx context.Context) (string, error) {
	return g.searchGitHub(ctx, "is:pr is:open author:@me")
}

func (g *Github) listReviewRequests(ctx context.Context) (string, error) {
	return g.searchGitHub(ctx, "is:pr is:open review-requested:@me")
}

func (g *Github) Name() string {
	return "github"
}

func (g *Github) ToolDefinition() *anthropic.ToolParam {
	return &anthropic.ToolParam{
		Name:        g.Name(),
		Description: anthropic.String("Interact with GitHub Pull Requests"),
		InputSchema: GenerateSchema[GithubInput](),
	}
}
