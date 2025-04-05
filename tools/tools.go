package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type ToolCall interface {
	Run(ctx context.Context, arguments json.RawMessage) (string, error)
	Name() string
	ToolDefinition() *anthropic.ToolParam
}

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
