package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type Linear struct {
	token string
}

func NewLinear(token string) *Linear {
	return &Linear{
		token: token,
	}
}

type graphqlRequest struct {
	Query string `json:"query"`
}

type teamResponse struct {
	Data struct {
		Teams struct {
			Nodes []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"nodes"`
		} `json:"teams"`
	} `json:"data"`
}

type issueResponse struct {
	Data struct {
		Teams struct {
			Nodes []struct {
				Name   string `json:"name"`
				Issues struct {
					Nodes []struct {
						ID    string `json:"id"`
						Title string `json:"title"`
						State struct {
							Name  string `json:"name"`
							Color string `json:"color"`
							Type  string `json:"type"`
						} `json:"state"`
						Priority int    `json:"priority"`
						URL      string `json:"url"`
						Assignee struct {
							Name string `json:"name"`
						} `json:"assignee"`
						CreatedAt string `json:"createdAt"`
					} `json:"nodes"`
				} `json:"issues"`
			} `json:"nodes"`
		} `json:"teams"`
	} `json:"data"`
}

func (l *Linear) makeRequest(ctx context.Context, query string) ([]byte, error) {
	reqBody, err := json.Marshal(graphqlRequest{Query: query})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.linear.app/graphql", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", l.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Linear API returned status %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

func (l *Linear) Run(ctx context.Context, arguments json.RawMessage) (string, error) {
	var inputs LinearToolInputs
	if err := json.Unmarshal(arguments, &inputs); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	switch inputs.Action {
	case "get_my_teams_in_review_issues":
		return l.getMyTeamsInReviewIssues(ctx, inputs.Teams, inputs.Name)
	case "get_my_issues":
		return l.getMyIssues(ctx, inputs.Name)
	default:
		return "", fmt.Errorf("invalid action: %s", inputs.Action)
	}
}

func (l *Linear) getMyTeamsInReviewIssues(ctx context.Context, teams []string, name string) (string, error) {
	fmt.Println("Getting my teams in review issues.")

	teamKeys := make([]string, 0)
	for _, team := range teams {
		teamKeys = append(teamKeys, fmt.Sprintf(`"%s"`, team))
	}

	issuesQuery := fmt.Sprintf(`
	query {
		issues(
			first: 40,
			orderBy: updatedAt, 
			filter: { 
				team: { key: { in: [%s] } },
				state: { name: { in: "In Review" }}
				assignee: { name: {nin: "%s"}}
			}
		) {
			nodes {
				identifier
				title
				team {
					name
				}
				state {
					name
					type
				}
				assignee {
					name
				}
				url
			}
		}
	}
	`, strings.Join(teamKeys, ", "), name)

	issuesBody, err := l.makeRequest(ctx, issuesQuery)
	if err != nil {
		return "", err
	}

	return string(issuesBody), nil
}

func (l *Linear) getMyIssues(ctx context.Context, name string) (string, error) {
	fmt.Println("Getting my issues.")

	issuesQuery := fmt.Sprintf(`
	query {
		issues(
			first: 40,
			orderBy: updatedAt,	
			filter: { 
				assignee: { name: {in: "%s"}}
			}
		) {
			nodes {
				identifier
				title
				team {
					name
				}
				state {
					name
					type
				}
				priority
				assignee {
					name
				}
				url
			}
		}
	}
	`, name)

	issuesBody, err := l.makeRequest(ctx, issuesQuery)
	if err != nil {
		return "", err
	}

	return string(issuesBody), nil
}

func (l *Linear) Name() string {
	return "linear"
}

func (l *Linear) ToolDefinition() *anthropic.ToolParam {
	return &anthropic.ToolParam{
		Name:        l.Name(),
		Description: anthropic.String("Get status of all tasks for your team in Linear"),
		InputSchema: GenerateSchema[LinearToolInputs](),
	}
}

type LinearToolInputs struct {
	Action string   `json:"action" jsonschema_description:"The action to perform (get_my_teams_in_review_issues, get_my_issues)"`
	Teams  []string `json:"teams" jsonschema_description:"The teams to get issues from"`
	Name   string   `json:"name" jsonschema_description:"Your name"`
}
