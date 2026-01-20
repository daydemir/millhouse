package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// VerificationResult contains git verification outcomes
type VerificationResult struct {
	CommitExists       bool
	CommitReachable    bool     // Is commit in current branch history?
	FilesMatch         []string // Files that match claims
	FilesMissing       []string // Claimed files not in commit
	UnstagedChanges    []string // Unstaged changes in repo
	UncommittedChanges bool     // Are there uncommitted changes?
	RemoteStatus       string   // ahead/behind/up-to-date
	Errors             []string
}

// VerifyCommitExists checks if a commit SHA exists in the repository
func VerifyCommitExists(basePath string, commitSHA string) (bool, error) {
	cmd := exec.Command("git", "cat-file", "-t", commitSHA)
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return false, nil // Commit doesn't exist (not an error, just false)
	}
	return strings.TrimSpace(string(output)) == "commit", nil
}

// VerifyCommitFiles checks if a commit contains the claimed files
func VerifyCommitFiles(basePath string, commitSHA string, files []string) (matches []string, missing []string, err error) {
	// Get list of files in commit
	cmd := exec.Command("git", "show", "--name-only", "--format=", commitSHA)
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commit files: %w", err)
	}

	commitFiles := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			commitFiles[line] = true
		}
	}

	// Check each claimed file
	for _, file := range files {
		if commitFiles[file] {
			matches = append(matches, file)
		} else {
			missing = append(missing, file)
		}
	}

	return matches, missing, nil
}

// CheckWorkingTreeClean verifies no unstaged or uncommitted changes
func CheckWorkingTreeClean(basePath string) (clean bool, changes []string, err error) {
	// Check for unstaged and uncommitted changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		return false, nil, fmt.Errorf("failed to check working tree: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return true, nil, nil
	}

	// Parse changes
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			changes = append(changes, line)
		}
	}

	return false, changes, nil
}

// CheckRemoteStatus checks if commits are pushed to remote
func CheckRemoteStatus(basePath string) (status string, err error) {
	// Check if there are unpushed commits
	cmd := exec.Command("git", "rev-list", "@{u}..HEAD")
	cmd.Dir = basePath
	output, err := cmd.Output()
	if err != nil {
		// If there's no upstream branch, return specific message
		return "no-upstream", nil
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return "up-to-date", nil
	}

	// Count unpushed commits
	unpushedCommits := len(strings.Split(outputStr, "\n"))
	return fmt.Sprintf("ahead-%d", unpushedCommits), nil
}

// VerifyEvidence checks git state against evidence claims
func VerifyEvidence(basePath string, commitSHA string, claimedFiles []string) (*VerificationResult, error) {
	result := &VerificationResult{}

	// 1. Verify commit exists
	exists, err := VerifyCommitExists(basePath, commitSHA)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error checking commit: %v", err))
		return result, err
	}
	result.CommitExists = exists

	if !exists {
		result.Errors = append(result.Errors, fmt.Sprintf("Commit %s does not exist", commitSHA))
		return result, nil
	}

	// 2. Verify commit contains claimed files
	matches, missing, err := VerifyCommitFiles(basePath, commitSHA, claimedFiles)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error checking commit files: %v", err))
		return result, err
	}
	result.FilesMatch = matches
	result.FilesMissing = missing

	// 3. Check working tree cleanliness
	clean, changes, err := CheckWorkingTreeClean(basePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error checking working tree: %v", err))
		return result, err
	}
	result.UncommittedChanges = !clean
	result.UnstagedChanges = changes

	// 4. Check remote status
	remoteStatus, err := CheckRemoteStatus(basePath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Error checking remote status: %v", err))
		return result, err
	}
	result.RemoteStatus = remoteStatus

	return result, nil
}

// IsVerified returns true if all verification checks passed
func (r *VerificationResult) IsVerified() bool {
	return r.CommitExists &&
		len(r.FilesMissing) == 0 &&
		!r.UncommittedChanges &&
		len(r.Errors) == 0
}

// GetErrorSummary returns a human-readable summary of verification failures
func (r *VerificationResult) GetErrorSummary() string {
	var issues []string

	if !r.CommitExists {
		issues = append(issues, "Commit does not exist (phantom commit)")
	}

	if len(r.FilesMissing) > 0 {
		issues = append(issues, fmt.Sprintf("Missing files in commit: %s", strings.Join(r.FilesMissing, ", ")))
	}

	if r.UncommittedChanges {
		issues = append(issues, fmt.Sprintf("Uncommitted changes detected: %d files", len(r.UnstagedChanges)))
	}

	if len(r.Errors) > 0 {
		issues = append(issues, fmt.Sprintf("Verification errors: %s", strings.Join(r.Errors, "; ")))
	}

	if len(issues) == 0 {
		return "All verification checks passed"
	}

	return strings.Join(issues, "; ")
}
