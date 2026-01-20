package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "millhouse-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git for testing
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to config git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to config git name: %v", err)
	}

	return tmpDir, cleanup
}

// createTestCommit creates a commit with specified files
func createTestCommit(t *testing.T, repoPath string, files []string, message string) string {
	// Create and add files
	for _, file := range files {
		filePath := filepath.Join(repoPath, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// Git add
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	// Git commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Get commit SHA
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit SHA: %v", err)
	}

	return string(output[:7]) // Return short SHA
}

func TestVerifyCommitExists(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a test commit
	commitSHA := createTestCommit(t, repo, []string{"test.txt"}, "Initial commit")

	tests := []struct {
		name      string
		commitSHA string
		want      bool
	}{
		{
			name:      "Valid commit exists",
			commitSHA: commitSHA,
			want:      true,
		},
		{
			name:      "Invalid commit does not exist",
			commitSHA: "abc123f",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VerifyCommitExists(repo, tt.commitSHA)
			if err != nil {
				t.Errorf("VerifyCommitExists() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("VerifyCommitExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyCommitFiles(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a commit with specific files
	files := []string{"src/main.go", "README.md"}
	commitSHA := createTestCommit(t, repo, files, "Add files")

	tests := []struct {
		name         string
		claimedFiles []string
		wantMatches  []string
		wantMissing  []string
	}{
		{
			name:         "All files match",
			claimedFiles: []string{"src/main.go", "README.md"},
			wantMatches:  []string{"src/main.go", "README.md"},
			wantMissing:  nil,
		},
		{
			name:         "Some files missing",
			claimedFiles: []string{"src/main.go", "missing.txt"},
			wantMatches:  []string{"src/main.go"},
			wantMissing:  []string{"missing.txt"},
		},
		{
			name:         "All files missing",
			claimedFiles: []string{"missing1.txt", "missing2.txt"},
			wantMatches:  nil,
			wantMissing:  []string{"missing1.txt", "missing2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches, missing, err := VerifyCommitFiles(repo, commitSHA, tt.claimedFiles)
			if err != nil {
				t.Errorf("VerifyCommitFiles() error = %v", err)
				return
			}

			if len(matches) != len(tt.wantMatches) {
				t.Errorf("VerifyCommitFiles() matches = %v, want %v", matches, tt.wantMatches)
			}

			if len(missing) != len(tt.wantMissing) {
				t.Errorf("VerifyCommitFiles() missing = %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func TestCheckWorkingTreeClean(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create initial commit
	createTestCommit(t, repo, []string{"initial.txt"}, "Initial commit")

	t.Run("Clean working tree", func(t *testing.T) {
		clean, changes, err := CheckWorkingTreeClean(repo)
		if err != nil {
			t.Errorf("CheckWorkingTreeClean() error = %v", err)
			return
		}
		if !clean {
			t.Errorf("CheckWorkingTreeClean() clean = false, want true")
		}
		if len(changes) != 0 {
			t.Errorf("CheckWorkingTreeClean() changes = %v, want empty", changes)
		}
	})

	t.Run("Dirty working tree with unstaged changes", func(t *testing.T) {
		// Create a new file without staging
		testFile := filepath.Join(repo, "unstaged.txt")
		if err := os.WriteFile(testFile, []byte("unstaged content"), 0644); err != nil {
			t.Fatalf("Failed to create unstaged file: %v", err)
		}

		clean, changes, err := CheckWorkingTreeClean(repo)
		if err != nil {
			t.Errorf("CheckWorkingTreeClean() error = %v", err)
			return
		}
		if clean {
			t.Errorf("CheckWorkingTreeClean() clean = true, want false")
		}
		if len(changes) == 0 {
			t.Errorf("CheckWorkingTreeClean() changes = empty, want non-empty")
		}
	})
}

func TestVerifyEvidence(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a commit with test files
	files := []string{"src/main.go", "README.md"}
	commitSHA := createTestCommit(t, repo, files, "Test commit")

	t.Run("Valid evidence passes all checks", func(t *testing.T) {
		result, err := VerifyEvidence(repo, commitSHA, files)
		if err != nil {
			t.Errorf("VerifyEvidence() error = %v", err)
			return
		}

		if !result.IsVerified() {
			t.Errorf("VerifyEvidence() should pass, got errors: %s", result.GetErrorSummary())
		}

		if !result.CommitExists {
			t.Errorf("Expected commit to exist")
		}

		if len(result.FilesMissing) > 0 {
			t.Errorf("Expected no missing files, got: %v", result.FilesMissing)
		}
	})

	t.Run("Phantom commit fails verification", func(t *testing.T) {
		result, err := VerifyEvidence(repo, "phantom123", files)
		if err != nil {
			t.Errorf("VerifyEvidence() error = %v", err)
			return
		}

		if result.IsVerified() {
			t.Errorf("Phantom commit should fail verification")
		}

		if result.CommitExists {
			t.Errorf("Phantom commit should not exist")
		}
	})

	t.Run("Missing files fail verification", func(t *testing.T) {
		claimedFiles := []string{"src/main.go", "missing.txt"}
		result, err := VerifyEvidence(repo, commitSHA, claimedFiles)
		if err != nil {
			t.Errorf("VerifyEvidence() error = %v", err)
			return
		}

		if result.IsVerified() {
			t.Errorf("Should fail with missing files")
		}

		if len(result.FilesMissing) == 0 {
			t.Errorf("Expected missing files")
		}
	})
}

func TestVerificationResultErrorSummary(t *testing.T) {
	tests := []struct {
		name   string
		result VerificationResult
		want   string
	}{
		{
			name: "All checks passed",
			result: VerificationResult{
				CommitExists:       true,
				FilesMatch:         []string{"file1.go"},
				FilesMissing:       []string{},
				UncommittedChanges: false,
			},
			want: "All verification checks passed",
		},
		{
			name: "Phantom commit",
			result: VerificationResult{
				CommitExists: false,
			},
			want: "Commit does not exist (phantom commit)",
		},
		{
			name: "Missing files",
			result: VerificationResult{
				CommitExists: true,
				FilesMissing: []string{"missing1.txt", "missing2.txt"},
			},
			want: "Missing files in commit: missing1.txt, missing2.txt",
		},
		{
			name: "Uncommitted changes",
			result: VerificationResult{
				CommitExists:       true,
				UncommittedChanges: true,
				UnstagedChanges:    []string{"?? file1.txt", "M file2.txt"},
			},
			want: "Uncommitted changes detected: 2 files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.GetErrorSummary()
			if got != tt.want {
				t.Errorf("GetErrorSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
