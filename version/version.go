package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ChangeType string

const (
	ChangeTypeUser   ChangeType = "user"
	ChangeTypeAgent  ChangeType = "agent"
	ChangeTypeSystem ChangeType = "system"
)

type Change struct {
	Timestamp time.Time  `json:"timestamp"`
	Type      ChangeType `json:"type"`
	Content   string     `json:"content"`
	Author    string     `json:"author"`
}

type VersionHistory struct {
	Changes []Change `json:"changes"`
}

func NewVersionHistory() *VersionHistory {
	return &VersionHistory{
		Changes: make([]Change, 0),
	}
}

func (vh *VersionHistory) AddChange(content string, changeType ChangeType, author string) {
	change := Change{
		Timestamp: time.Now(),
		Type:      changeType,
		Content:   content,
		Author:    author,
	}
	vh.Changes = append(vh.Changes, change)
}

func (vh *VersionHistory) GetLatestContent() string {
	if len(vh.Changes) == 0 {
		return ""
	}
	return vh.Changes[len(vh.Changes)-1].Content
}

func (vh *VersionHistory) SaveToFile(filePath string) error {
	data, err := json.MarshalIndent(vh, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version history: %v", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write version history: %v", err)
	}

	return nil
}

func LoadFromFile(filePath string) (*VersionHistory, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewVersionHistory(), nil
		}
		return nil, fmt.Errorf("failed to read version history: %v", err)
	}

	var vh VersionHistory
	if err := json.Unmarshal(data, &vh); err != nil {
		return nil, fmt.Errorf("failed to unmarshal version history: %v", err)
	}

	return &vh, nil
}

func (vh *VersionHistory) GetChangesByType(changeType ChangeType) []Change {
	var changes []Change
	for _, change := range vh.Changes {
		if change.Type == changeType {
			changes = append(changes, change)
		}
	}
	return changes
}

func (vh *VersionHistory) GetLastUserChange() *Change {
	for i := len(vh.Changes) - 1; i >= 0; i-- {
		if vh.Changes[i].Type == ChangeTypeUser {
			return &vh.Changes[i]
		}
	}
	return nil
}
