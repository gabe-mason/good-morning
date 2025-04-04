package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
	ollama "github.com/ollama/ollama/api"
)

type ToolCalls []ToolCall

func (o ToolCalls) GetTool(name string) (ToolCall, error) {
	for _, tool := range o {
		if tool.Name() == name {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", name)
}

func (o ToolCalls) GetToolDefinitions() []anthropic.ToolUnionParam {
	tools := make([]anthropic.ToolUnionParam, len(o))
	for i, toolParam := range o {
		tools[i] = anthropic.ToolUnionParam{OfTool: toolParam.ToolDefinition()}
	}
	return tools
}

type ToolCall interface {
	Run(ctx context.Context, arguments json.RawMessage) (string, error)
	Name() string
	ToolDefinition() *anthropic.ToolParam
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

type InvalidToolArgumentsError struct {
	ToolName string
	Message  string
}

func (e *InvalidToolArgumentsError) Error() string {
	return fmt.Sprintf("invalid arguments for tool %s: %s", e.ToolName, e.Message)
}

type CurrentWeather struct {
}

func (c *CurrentWeather) Run(arguments ollama.ToolCallFunctionArguments) (string, error) {
	if location, ok := arguments["location"].(string); ok {
		return "The weather in " + location + " is sunny", nil
	}
	return "not found, try again", nil
}

func (c *CurrentWeather) Name() string {
	return "getCurrentWeather"
}

func (c *CurrentWeather) ToolDefinition() ollama.Tool {
	return ollama.Tool{
		Type: "function",
		Function: ollama.ToolFunction{
			Name:        "getCurrentWeather",
			Description: "Get the current weather",
			Parameters: struct {
				Type       string   "json:\"type\""
				Required   []string "json:\"required\""
				Properties map[string]struct {
					Type        string   "json:\"type\""
					Description string   "json:\"description\""
					Enum        []string "json:\"enum,omitempty\""
				} "json:\"properties\""
			}{
				Type:     "object",
				Required: []string{"location", "unit"},
				Properties: map[string]struct {
					Type        string   "json:\"type\""
					Description string   "json:\"description\""
					Enum        []string "json:\"enum,omitempty\""
				}{
					"location": {
						Type:        "string",
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: "string",
						Enum: []string{"fahrenheit", "celsius"},
					},
				},
			},
		},
	}
}
