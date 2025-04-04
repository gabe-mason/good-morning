package agent

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	ollama "github.com/ollama/ollama/api"
)

type LlamaTokenCounter struct {
	// Llama uses a BPE tokenizer similar to GPT-2
	// We'll use a simplified version that approximates Llama's tokenization
}

func NewLlamaTokenCounter() *LlamaTokenCounter {
	return &LlamaTokenCounter{}
}

func (ltc *LlamaTokenCounter) CountTokens(message ollama.Message) int {
	// Base tokens for message structure
	tokens := 3 // Start token, role token, content token

	// Count tokens for content using Llama's approximate tokenization
	tokens += ltc.countContentTokens(message.Content)

	// Count tokens for role
	tokens += ltc.countContentTokens(message.Role)

	// Count tokens for tool calls if present
	if len(message.ToolCalls) > 0 {
		for _, toolCall := range message.ToolCalls {
			tokens += 3 // Function start, name, arguments
			tokens += ltc.countContentTokens(toolCall.Function.Name)

			// Count argument tokens
			if args, err := json.Marshal(toolCall.Function.Arguments); err == nil {
				tokens += ltc.countContentTokens(string(args))
			}
		}
	}

	return tokens
}

func (ltc *LlamaTokenCounter) countContentTokens(content string) int {
	// Llama's tokenizer is based on BPE with a vocabulary of ~32k tokens
	// This is a simplified approximation:

	// Split into words and count tokens
	words := strings.Fields(content)
	tokenCount := 0

	for _, word := range words {
		// Count characters (approximate)
		charCount := utf8.RuneCountInString(word)

		// Llama's tokenizer typically splits words into subwords
		// A rough approximation: 1 token per 3-4 characters
		tokenCount += (charCount + 3) / 4

		// Add token for space after word
		tokenCount++
	}

	return tokenCount
}

// // Package-level function for estimating total tokens
// func estimateTotalTokens(messages []ollama.Message) int {
// 	counter := NewLlamaTokenCounter()
// 	total := 0
// 	for _, msg := range messages {
// 		total += counter.CountTokens(msg)
// 	}
// 	return total
// }
