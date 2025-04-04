package agent

import (
	"context"
	"fmt"
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
	config         *config.Config
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
		config:        config,
	}
}

func (a *Agent) GenerateDailySummary(ctx context.Context) (string, error) {
	a.contextManager.AppendUserMessage("The current date is " + time.Now().Format("2006-01-02"))
	a.contextManager.AppendUserMessage("My name is " + a.config.MyName + " and I'm an engineer in teams " + a.config.LinearTeams)
	a.contextManager.AppendUserMessage("What's the plan for today? Check my calendar for meetings and Linear for any issues I need to review or work on. Include Zoom links or Linear links if they exist.  I like emojis, please use them.")
	a.contextManager.AppendUserMessage("Create a section for each meeting I have, include a note of the people in attendance and the topic of the meeting, with some space for notes.")
	a.contextManager.AppendUserMessage("Create a section for each issue I need to review for my team, is is everyone except me, include a note of the title, author, and priority of the issue with a link to the issue, do not redact.")
	a.contextManager.AppendUserMessage("Create a section for each thing I need to do, include a note of the title, author, and priority of the issue with a link to the issue.")
	a.contextManager.AppendUserMessage("Create a markdown formatted with the following gist")
	a.contextManager.AppendUserMessage(`Start each file with some interesting ASCII art max 8 x 8 characters.
{ascii art}
	# Good Morning {name}!
{tell me a joke}
## Calendar 📅

### {time} | {meeting title}
- **Attendees**: {attendees}
- **Topic**: {meeting topic}
- **Zoom**: {zoom link}

#### Notes:
- 

## Things I need to review 👀

n items need review from my team:

1. [task identifier](link) - title (assigned to)

## Things I need to do ✅

Active issues assigned to you:

**High Priority**:
- (emoji representing priority) [task identifier](link) - title(status)

**In Progress**:
- (emoji representing priority) [task identifier](link) - title(status)

**To Do**:
- (emoji representing priority) [task identifier](link) - title(status)
## Suggestions 💡
-- Add suggestions for my day here
`)
	myPullRequests, err := a.callModel(ctx)
	if err != nil {
		return "", err
	}

	return myPullRequests, nil
}

func (a *Agent) callModel(ctx context.Context) (string, error) {
	for a.contextManager.HasNewMessages() {
		response, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			MaxTokens: 8000,
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
				fmt.Println(block.Text)
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
	fmt.Println("Going to have a chin wag with " + block.Name + ".")
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
	fmt.Println(block.Name + " has answered my questions.")
	return anthropic.NewToolResultBlock(block.ID, toolResult, false), nil
}
