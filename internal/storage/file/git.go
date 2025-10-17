package file

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/leifj/go-raid/internal/models"
	"github.com/leifj/go-raid/internal/storage"
)

func init() {
	// Register git storage factory
	storage.RegisterFactory(storage.StorageTypeFileGit, func(cfg interface{}) (storage.Repository, error) {
		fileCfg, ok := cfg.(*storage.FileConfig)
		if !ok || fileCfg == nil {
			fileCfg = &storage.FileConfig{
				DataDir:       "./data",
				GitEnabled:    true,
				GitAutoCommit: true,
			}
		}
		return NewGitStorage(&GitConfig{
			FileConfig:  &Config{DataDir: fileCfg.DataDir},
			Enabled:     true,
			AutoCommit:  fileCfg.GitAutoCommit,
			AuthorName:  fileCfg.GitAuthorName,
			AuthorEmail: fileCfg.GitAuthorEmail,
		})
	})
}

// GitStorage wraps FileStorage and adds git commit functionality
type GitStorage struct {
	*FileStorage
	gitEnabled  bool
	autoCommit  bool
	authorName  string
	authorEmail string
}

// GitConfig holds configuration for git-enabled storage
type GitConfig struct {
	FileConfig  *Config
	Enabled     bool
	AutoCommit  bool
	AuthorName  string
	AuthorEmail string
}

// NewGitStorage creates a new git-enabled file storage
func NewGitStorage(cfg *GitConfig) (*GitStorage, error) {
	// Create underlying file storage
	fs, err := New(cfg.FileConfig)
	if err != nil {
		return nil, err
	}

	gs := &GitStorage{
		FileStorage: fs,
		gitEnabled:  cfg.Enabled,
		autoCommit:  cfg.AutoCommit,
		authorName:  cfg.AuthorName,
		authorEmail: cfg.AuthorEmail,
	}

	// Set defaults
	if gs.authorName == "" {
		gs.authorName = "RAiD System"
	}
	if gs.authorEmail == "" {
		gs.authorEmail = "raid@example.org"
	}

	// Initialize git repository if enabled
	if gs.gitEnabled {
		if err := gs.initGitRepo(); err != nil {
			return nil, fmt.Errorf("failed to initialize git repository: %w", err)
		}
	}

	return gs, nil
}

// CreateRAiD mints a new RAiD and commits to git
func (gs *GitStorage) CreateRAiD(ctx context.Context, raid *models.RAiD) (*models.RAiD, error) {
	result, err := gs.FileStorage.CreateRAiD(ctx, raid)
	if err != nil {
		return nil, err
	}

	if gs.gitEnabled && gs.autoCommit {
		prefix, suffix, _ := parseRAiDIdentifier(result.Identifier.ID)
		commitMsg := fmt.Sprintf("Create RAiD %s/%s", prefix, suffix)
		if err := gs.gitCommit(commitMsg); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return result, nil
}

// UpdateRAiD updates a RAiD and commits to git
func (gs *GitStorage) UpdateRAiD(ctx context.Context, prefix, suffix string, raid *models.RAiD) (*models.RAiD, error) {
	result, err := gs.FileStorage.UpdateRAiD(ctx, prefix, suffix, raid)
	if err != nil {
		return nil, err
	}

	if gs.gitEnabled && gs.autoCommit {
		commitMsg := fmt.Sprintf("Update RAiD %s/%s to version %d", prefix, suffix, result.Identifier.Version)
		if err := gs.gitCommit(commitMsg); err != nil {
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return result, nil
}

// DeleteRAiD deletes a RAiD and commits to git
func (gs *GitStorage) DeleteRAiD(ctx context.Context, prefix, suffix string) error {
	if err := gs.FileStorage.DeleteRAiD(ctx, prefix, suffix); err != nil {
		return err
	}

	if gs.gitEnabled && gs.autoCommit {
		commitMsg := fmt.Sprintf("Delete RAiD %s/%s", prefix, suffix)
		if err := gs.gitCommit(commitMsg); err != nil {
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return nil
}

// CreateServicePoint creates a service point and commits to git
func (gs *GitStorage) CreateServicePoint(ctx context.Context, sp *models.ServicePoint) (*models.ServicePoint, error) {
	result, err := gs.FileStorage.CreateServicePoint(ctx, sp)
	if err != nil {
		return nil, err
	}

	if gs.gitEnabled && gs.autoCommit {
		commitMsg := fmt.Sprintf("Create service point %d (%s)", result.ID, result.Name)
		if err := gs.gitCommit(commitMsg); err != nil {
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return result, nil
}

// UpdateServicePoint updates a service point and commits to git
func (gs *GitStorage) UpdateServicePoint(ctx context.Context, id int64, sp *models.ServicePoint) (*models.ServicePoint, error) {
	result, err := gs.FileStorage.UpdateServicePoint(ctx, id, sp)
	if err != nil {
		return nil, err
	}

	if gs.gitEnabled && gs.autoCommit {
		commitMsg := fmt.Sprintf("Update service point %d (%s)", id, result.Name)
		if err := gs.gitCommit(commitMsg); err != nil {
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return result, nil
}

// DeleteServicePoint deletes a service point and commits to git
func (gs *GitStorage) DeleteServicePoint(ctx context.Context, id int64) error {
	if err := gs.FileStorage.DeleteServicePoint(ctx, id); err != nil {
		return err
	}

	if gs.gitEnabled && gs.autoCommit {
		commitMsg := fmt.Sprintf("Delete service point %d", id)
		if err := gs.gitCommit(commitMsg); err != nil {
			fmt.Printf("Git commit failed: %v\n", err)
		}
	}

	return nil
}

// GetGitLog retrieves the git log for a specific file
func (gs *GitStorage) GetGitLog(prefix, suffix string) ([]GitCommit, error) {
	if !gs.gitEnabled {
		return nil, fmt.Errorf("git is not enabled")
	}

	filePath := filepath.Join("raids", sanitizePath(prefix), sanitizePath(suffix)+".json")
	cmd := exec.Command("git", "-C", gs.dataDir, "log", "--pretty=format:%H|%an|%ae|%at|%s", "--", filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	commits := make([]GitCommit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) != 5 {
			continue
		}

		var timestamp int64
		fmt.Sscanf(parts[3], "%d", &timestamp)

		commits = append(commits, GitCommit{
			Hash:      parts[0],
			Author:    parts[1],
			Email:     parts[2],
			Timestamp: time.Unix(timestamp, 0),
			Message:   parts[4],
		})
	}

	return commits, nil
}

// Git helper methods

func (gs *GitStorage) initGitRepo() error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	// Check if already a git repository
	checkCmd := exec.Command("git", "-C", gs.dataDir, "rev-parse", "--git-dir")
	if err := checkCmd.Run(); err == nil {
		// Already a git repository
		return nil
	}

	// Initialize new git repository
	initCmd := exec.Command("git", "-C", gs.dataDir, "init")
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to init git repository: %w", err)
	}

	// Configure git
	gs.runGitCommand("config", "user.name", gs.authorName)
	gs.runGitCommand("config", "user.email", gs.authorEmail)

	// Create initial commit
	gs.runGitCommand("commit", "--allow-empty", "-m", "Initial commit")

	return nil
}

func (gs *GitStorage) gitCommit(message string) error {
	// Add all changes
	if err := gs.runGitCommand("add", "-A"); err != nil {
		return err
	}

	// Commit
	if err := gs.runGitCommand("commit", "-m", message, "--author", fmt.Sprintf("%s <%s>", gs.authorName, gs.authorEmail)); err != nil {
		// Check if it's a "nothing to commit" error
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil
		}
		return err
	}

	return nil
}

func (gs *GitStorage) runGitCommand(args ...string) error {
	fullArgs := append([]string{"-C", gs.dataDir}, args...)
	cmd := exec.Command("git", fullArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git command failed: %w, output: %s", err, string(output))
	}

	return nil
}

// GitCommit represents a git commit
type GitCommit struct {
	Hash      string
	Author    string
	Email     string
	Timestamp time.Time
	Message   string
}

// Verify GitStorage implements storage.Repository
var _ storage.Repository = (*GitStorage)(nil)
