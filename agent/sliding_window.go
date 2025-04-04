package agent

type SlidingWindow struct {
	maxMessages int
	messages    []PrioritizedMessage
}

func (sw *SlidingWindow) Add(message PrioritizedMessage) {
	if len(sw.messages) >= sw.maxMessages {
		// Remove lowest priority message
		sw.removeLowestPriority()
	}
	sw.messages = append(sw.messages, message)
}

func (sw *SlidingWindow) removeLowestPriority() {
	// Find the lowest priority message
	lowestPriority := PriorityLow
	for _, msg := range sw.messages {
		if msg.Priority < lowestPriority {
			lowestPriority = msg.Priority
		}
	}
	for i, msg := range sw.messages {
		if msg.Priority == lowestPriority {
			sw.messages = append(sw.messages[:i], sw.messages[i+1:]...)
			break
		}
	}
}
