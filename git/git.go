package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type GitManager struct {
	repoPath string
}

func NewGitManager(repoPath string) (*GitManager, error) {
	// Initialize git repository if it doesn't exist
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to initialize git repository: %v", err)
		}

		// Create .gitignore
		gitignore := `# System files
.DS_Store
.env

# IDE files
.idea/
.vscode/

# Log files
*.log
`
		if err := os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte(gitignore), 0644); err != nil {
			return nil, fmt.Errorf("failed to create .gitignore: %v", err)
		}
	}

	return &GitManager{
		repoPath: repoPath,
	}, nil
}

func (gm *GitManager) CommitChanges(message string, author string) error {
	// Add all changes
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = gm.repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %v", err)
	}

	// Commit changes
	cmd = exec.Command("git", "commit", "-m", message, "--author", author)
	cmd.Dir = gm.repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	return nil
}

func (gm *GitManager) GetLastUserCommit() (string, error) {
	// Get the last commit by the user
	cmd := exec.Command("git", "log", "--author=user", "-1", "--pretty=format:%H")
	cmd.Dir = gm.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last user commit: %v", err)
	}

	return string(output), nil
}

func (gm *GitManager) GetFileContentAtCommit(commitHash, filePath string) (string, error) {
	// Get file content at specific commit
	cmd := exec.Command("git", "show", commitHash+":"+filePath)
	cmd.Dir = gm.repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get file content at commit: %v", err)
	}

	return string(output), nil
}

func (gm *GitManager) GetCurrentFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filepath.Join(gm.repoPath, filePath))
	if err != nil {
		return "", fmt.Errorf("failed to read current file content: %v", err)
	}
	return string(content), nil
}
