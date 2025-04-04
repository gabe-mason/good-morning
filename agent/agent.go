package agent

import (
	"context"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gabe-mason/good-morning/config"
	"github.com/gabe-mason/good-morning/tools"
)

type Agent struct {
	contextManager *ContextManager
	systemPrompt   string
	client         anthropic.Client
	tools          tools.ToolCalls
	maxRetries     int
	state          AgentState
	contextWindow  int
	modelName      string
}

type AgentState struct {
	metadata           map[string]interface{}
	ConversationActive bool
}

func NewAgent(client anthropic.Client, tools tools.ToolCalls, config *config.Config) *Agent {
	contextWindow := 4000 // Default context window size
	contextManager := CreateContextManager(
		tools,
		contextWindow,
		config.GetContextManagerLocation(),
	)

	return &Agent{
		systemPrompt: `You are an AI agent that can use tools to produce a daily summary of my day.
		Be concise and to the point, do not include any other text.
		You must only return markdown formatted text.`,
		contextManager: contextManager,
		client:         client,
		tools:          tools,
		maxRetries:     3,
		state: AgentState{
			ConversationActive: true,
			metadata:           make(map[string]interface{}),
		},
		contextWindow: contextWindow,
		modelName:     anthropic.ModelClaude3_5SonnetLatest,
	}
}

func (a *Agent) GenerateDailySummary(ctx context.Context) (string, error) {
	myPullRequests, err := a.getGithubPRs(ctx)
	if err != nil {
		return "", err
	}

	return myPullRequests, nil
}

func (a *Agent) getCalendarEvents(ctx context.Context) (string, error) {
	a.contextManager.AppendUserMessage("What does my calendar look like for today? Create a markdown formatted table response. If the meetings have Zoom links then include them in the table.")
	a.contextManager.AppendUserMessage("The current date is " + time.Now().Format("2006-01-02"))
	response, err := a.callModel(ctx)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (a *Agent) getGithubPRs(ctx context.Context) (string, error) {
	a.contextManager.AppendUserMessage("What are my open pull requests? Create a markdown formatted table response. If the pull requests have links then include them in the table.")
	response, err := a.callModel(ctx)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (a *Agent) callModel(ctx context.Context) (string, error) {
	for a.contextManager.HasNewMessages() {
		response, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			MaxTokens: 1024,
			Messages:  a.contextManager.GetMessages(),
			Model:     a.modelName,
			Tools:     a.tools.GetToolDefinitions(),
			System: []anthropic.TextBlockParam{
				{
					Type: "text",
					Text: a.systemPrompt,
				},
			},
		})
		a.contextManager.ClearNewMessages()
		if err != nil {
			return "", err
		}
		a.contextManager.AppendAssistantMessage(response.ToParam())
		responseText, err := a.processResponse(ctx, response)
		if err != nil {
			return "", err
		}
		if responseText != "" {
			return responseText, nil
		}
	}
	return "", nil
}

func (a *Agent) processResponse(ctx context.Context, response *anthropic.Message) (string, error) {
	toolResults := make([]anthropic.ContentBlockParamUnion, 0)
	for _, block := range response.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:
			if response.StopReason == "tool_use" {
				continue
			}
			return block.Text, nil
		case anthropic.ToolUseBlock:
			toolResponse, err := a.processToolUseBlock(ctx, &block)
			if err != nil {
				return "", err
			}
			toolResults = append(toolResults, toolResponse)
		}
	}
	a.contextManager.AppendToolResults(toolResults)
	return "", nil
}

func (a *Agent) processToolUseBlock(ctx context.Context, block *anthropic.ToolUseBlock) (anthropic.ContentBlockParamUnion, error) {
	input := block.Input
	tool, err := a.tools.GetTool(block.Name)
	if err != nil {
		return anthropic.NewToolResultBlock(block.ID, "This tool is not in the list of tools.", true), nil
	}
	toolResult, err := tool.Run(ctx, input)
	if err != nil {
		if invalidArgs, ok := err.(*tools.InvalidToolArgumentsError); ok {
			return anthropic.NewToolResultBlock(block.ID, invalidArgs.Error(), true), nil
		}
		return anthropic.NewToolResultBlock(block.ID, "An error occurred while running the tool.", true), nil
	}
	return anthropic.NewToolResultBlock(block.ID, toolResult, false), nil
}
