package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/gabe-mason/good-morning/tools"
)

type ContextManager struct {
	fileLocation   string
	messages       []anthropic.MessageParam
	contextWindow  int
	hasNewMessages bool
}

func CreateContextManager(tools tools.ToolCalls, contextWindow int, fileLocation string) *ContextManager {
	// Check if file exists, create if it doesn't
	if _, err := os.Stat(fileLocation); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		dir := filepath.Dir(fileLocation)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", dir, err)
		}

		// Create empty file
		file, err := os.Create(fileLocation)
		if err != nil {
			log.Printf("Error creating file %s: %v", fileLocation, err)
		} else {
			file.Close()
		}
	}

	// Create context manager
	cm := &ContextManager{
		messages:       []anthropic.MessageParam{},
		contextWindow:  contextWindow,
		hasNewMessages: false,
		fileLocation:   fileLocation,
	}

	return cm
}

func (ml *ContextManager) save() error {
	// Convert messages to JSON
	data, err := json.MarshalIndent(ml.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling messages: %v", err)
	}

	// Write to file
	if err := os.WriteFile(ml.fileLocation, data, 0644); err != nil {
		return fmt.Errorf("error writing messages to file: %v", err)
	}

	return nil
}

func (ml *ContextManager) AppendUserMessage(message string) {
	ml.append(anthropic.MessageParam{
		Role: anthropic.MessageParamRoleUser,
		Content: []anthropic.ContentBlockParamUnion{{
			OfRequestTextBlock: &anthropic.TextBlockParam{Text: message},
		}},
	})
	ml.hasNewMessages = true
}

func (ml *ContextManager) AppendToolResults(toolResults []anthropic.ContentBlockParamUnion) {
	ml.append(anthropic.MessageParam{
		Role:    anthropic.MessageParamRoleUser,
		Content: toolResults,
	})
	ml.hasNewMessages = true
}

func (ml *ContextManager) AppendAssistantMessage(message anthropic.MessageParam) {
	ml.append(message)
}

func (ml *ContextManager) AppendAssistantMessageAck(message anthropic.MessageParam) {
	ml.append(message)
	ml.hasNewMessages = true
}

func (ml *ContextManager) append(message anthropic.MessageParam) {
	ml.messages = append(ml.messages, message)
	ml.save()
}

func (ml *ContextManager) GetMessages() []anthropic.MessageParam {
	return ml.messages
}

func (ml *ContextManager) HasNewMessages() bool {
	return ml.hasNewMessages
}

func (ml *ContextManager) ClearNewMessages() {
	ml.hasNewMessages = false
}

func (ml *ContextManager) GetLastMessage() anthropic.MessageParam {
	return ml.messages[len(ml.messages)-1]
}
