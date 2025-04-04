package agent

import ollama "github.com/ollama/ollama/api"

type MessagePriority int

const (
	PrioritySystem MessagePriority = iota
	PriorityCritical
	PriorityNormal
	PriorityLow
)

type PrioritizedMessage struct {
	Message  ollama.Message
	Priority MessagePriority
}
