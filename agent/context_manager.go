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

	// Load existing messages
	// if err := cm.LoadMessages(); err != nil {
	// 	log.Printf("Error loading messages: %v", err)
	// }

	return cm
}

// func (ml *ContextManager) LoadMessages() error {
// 	data, err := os.ReadFile(ml.fileLocation)
// 	if err != nil {
// 		return fmt.Errorf("error reading messages from file: %v", err)
// 	}
// 	anthropic.Message
// 	json.Unmarshal(data, &ml.messages)

// 	return nil
// }

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

// func (ml *ContextManager) GetOldMessagesForSummarization() ([]anthropic.MessageParam, error) {
// 	if len(ml.messages) < 10 {
// 		return nil, fmt.Errorf("not enough messages to summarize")
// 	}

// 	// Get messages to summarize (all except system prompt and last 3 messages)
// 	return ml.messages[1 : len(ml.messages)-3], nil
// }

// func (ml *ContextManager) ManageContext(message anthropic.MessageParam) error {
// 	// Add the new message
// 	ml.AddMessage(message)

// 	// Check if we need to prune old messages
// 	totalTokens := ml.CountTokens(message)
// 	if totalTokens > ml.contextWindow {
// 		// Keep system prompt and last few messages
// 		ml.messages = append([]anthropic.MessageParam{ml.messages[0]}, ml.messages[len(ml.messages)-3:]...)
// 	}

// 	return nil
// }

// func (ml *ContextManager) CountTokens(message anthropic.MessageParam) int {
// 	// Base tokens for message structure
// 	tokens := 3 // Start token, role token, content token

// 	// Count tokens for content using Claude's approximate tokenization
// 	for _, content := range message.Content {
// 		if content.OfRequestTextBlock != nil {
// 			tokens += ml.countContentTokens(content.OfRequestTextBlock.Text)
// 		}
// 	}

// 	// Count tokens for role
// 	tokens += ml.countContentTokens(string(message.Role))

// 	return tokens
// }

// func (ml *ContextManager) countContentTokens(content string) int {
// 	// Claude's tokenizer is based on BPE with a vocabulary of ~100k tokens
// 	// This is a simplified approximation:

// 	// Split into words and count tokens
// 	words := strings.Fields(content)
// 	tokenCount := 0

// 	for _, word := range words {
// 		// Count characters (approximate)
// 		charCount := utf8.RuneCountInString(word)

// 		// Claude's tokenizer typically splits words into subwords
// 		// A rough approximation: 1 token per 3-4 characters
// 		tokenCount += (charCount + 3) / 4

// 		// Add token for space after word
// 		tokenCount++
// 	}

// 	return tokenCount
// }

func (ml *ContextManager) GetLastMessage() anthropic.MessageParam {
	return ml.messages[len(ml.messages)-1]
}

// func (ml *MessageLog) GetRespondFunc(expectation func(resp anthropic.Message) error) func(anthropic.Message) error {
// 	return func(resp anthropic.Message) error {
// 		// Add the response message
// 		ml.AddMessage(anthropic.MessageParam{
// 			Role:    anthropic.MessageParamRoleAssistant,
// 			Content: resp.Content,
// 		})

// 		// Check for tool calls
// 		if resp.ToolCalls != nil {
// 			for _, tc := range resp.ToolCalls {
// 				tool := ml.tools.GetTool(tc.Function.Name)
// 				if tool == nil {
// 					return fmt.Errorf("tool %s not found", tc.Function.Name)
// 				}

// 				// Execute tool
// 				toolResponse, err := tool.Run(tc.Function.Arguments)
// 				if err != nil {
// 					return err
// 				}

// 				// Add tool response
// 				ml.AddMessage(anthropic.MessageParam{
// 					Role: anthropic.MessageParamRoleAssistant,
// 					Content: []anthropic.ContentBlockParamUnion{
// 						{
// 							OfRequestTextBlock: &anthropic.TextBlockParam{
// 								Text: toolResponse,
// 							},
// 						},
// 					},
// 				})
// 			}
// 		}

// 		return expectation(resp)
// 	}
// }

// func (ml *ContextManager) ReplaceWithSummary(summary string) {
// 	// Create new message slice with proper capacity
// 	newMessages := make([]anthropic.MessageParam, 0, len(ml.messages))

// 	// Keep system prompt
// 	newMessages = append(newMessages, ml.messages[0])

// 	// Add summary
// 	newMessages = append(newMessages, anthropic.MessageParam{
// 		Role: anthropic.MessageParamRoleAssistant,
// 		Content: []anthropic.ContentBlockParamUnion{
// 			{
// 				OfRequestTextBlock: &anthropic.TextBlockParam{
// 					Text: summary,
// 				},
// 			},
// 		},
// 	})

// 	// Keep last few messages
// 	if len(ml.messages) > 3 {
// 		newMessages = append(newMessages, ml.messages[len(ml.messages)-3:]...)
// 	}

// 	ml.messages = newMessages
// }
