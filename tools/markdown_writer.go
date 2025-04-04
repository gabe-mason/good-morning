package tools

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
)

type MarkdownWriter struct {
	write chan<- Section
}

type Section struct {
	Title    string   `json:"title" jsonschema_description:"The title of the section e.g. My PRs"`
	Content  string   `json:"content" jsonschema_description:"The content of the section"`
	Comments []string `json:"comments" jsonschema_description:"The comments to write to the section"`
}

func NewMarkdownWriter(write chan<- Section) *MarkdownWriter {
	return &MarkdownWriter{
		write: write,
	}
}

type MarkdownWriterInputs struct {
	Sections []Section `json:"sections" jsonschema_description:"The sections to write to the markdown document"`
}

func (m *MarkdownWriter) Run(ctx context.Context, arguments json.RawMessage) (string, error) {

	var section Section
	if err := json.Unmarshal(arguments, &section); err != nil {
		return "", &InvalidToolArgumentsError{
			ToolName: m.Name(),
			Message:  "invalid JSON format",
		}
	}

	select {
	case m.write <- section:
		return "Section written successfully", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (m *MarkdownWriter) Name() string {
	return "markdown_writer"
}

func (m *MarkdownWriter) ToolDefinition() *anthropic.ToolParam {
	return &anthropic.ToolParam{
		Name:        m.Name(),
		Description: anthropic.String("Write a section to the markdown document"),
		InputSchema: GenerateSchema[MarkdownWriterInputs](),
	}
}
