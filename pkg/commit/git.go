package commit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/term"
)

type GitOperations struct {
	repo *git.Repository
}

type GitConfig struct {
	UserName   string
	UserEmail  string
	GPGSign    bool
	SigningKey string
	GPGProgram string
}

func NewGitOperations(repoPath string) (*GitOperations, error) {
	repo, err := git.PlainOpenWithOptions(repoPath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GitOperations{repo: repo}, nil
}

// GetConfig reads git configuration - fails if user.name or user.email not configured
func (g *GitOperations) GetConfig() (*GitConfig, error) {
	config := &GitConfig{
		GPGSign:    false,
		GPGProgram: "gpg",
	}

	// Get required user configuration
	userName := g.getConfigValue("user.name")
	if userName == "" {
		return nil, fmt.Errorf("git user.name not configured. Run: git config user.name \"Your Name\"")
	}
	config.UserName = userName

	userEmail := g.getConfigValue("user.email")
	if userEmail == "" {
		return nil, fmt.Errorf("git user.email not configured. Run: git config user.email \"your.email@example.com\"")
	}
	config.UserEmail = userEmail

	// Read optional GPG configuration
	if gpgSign := g.getConfigValue("commit.gpgsign"); gpgSign != "" {
		config.GPGSign = strings.ToLower(gpgSign) == "true"
	}
	if signingKey := g.getConfigValue("user.signingkey"); signingKey != "" {
		config.SigningKey = signingKey
	}
	if gpgProgram := g.getConfigValue("gpg.program"); gpgProgram != "" {
		config.GPGProgram = gpgProgram
	}

	return config, nil
}

// getConfigValue reads a specific git config value using git command
func (g *GitOperations) getConfigValue(key string) string {
	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (g *GitOperations) GetCurrentBranch() (string, error) {
	head, err := g.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()
	return branchName, nil
}

func (g *GitOperations) GetWorkingTreeStatus() (git.Status, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return status, nil
}

func (g *GitOperations) UnstageAll() error {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Single reset operation instead of per-file operations
	err = worktree.Reset(&git.ResetOptions{
		Mode: git.MixedReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	return nil
}

func (g *GitOperations) StageFiles(excludePatterns []string, includePatterns []string) ([]string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Optimization: if no patterns specified, use AddWithOptions for better performance
	if len(excludePatterns) == 0 && len(includePatterns) == 0 {
		return g.stageAllModified(worktree)
	}

	// If we have simple include patterns (glob-compatible), try to use AddGlob
	if len(excludePatterns) == 0 && len(includePatterns) == 1 && isSimpleGlobPattern(includePatterns[0]) {
		return g.stageWithGlob(worktree, includePatterns[0])
	}

	// Fall back to filtered staging for complex patterns
	return g.stageFiltered(worktree, excludePatterns, includePatterns)
}

// Fast path: stage all modified files
func (g *GitOperations) stageAllModified(worktree *git.Worktree) ([]string, error) {
	// Get status first to return the list of staged files
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var modifiedFiles []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree != git.Unmodified {
			modifiedFiles = append(modifiedFiles, file)
		}
	}

	if len(modifiedFiles) == 0 {
		return []string{}, nil
	}

	// Use AddWithOptions with All flag for better performance
	err = worktree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to stage all files: %w", err)
	}

	return modifiedFiles, nil
}

// Fast path: use glob patterns when possible
func (g *GitOperations) stageWithGlob(worktree *git.Worktree, pattern string) ([]string, error) {
	// Get status first to return the list of staged files
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var matchingFiles []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree == git.Unmodified {
			continue
		}
		if matched, _ := filepath.Match(pattern, file); matched {
			matchingFiles = append(matchingFiles, file)
		}
	}

	if len(matchingFiles) == 0 {
		return []string{}, nil
	}

	err = worktree.AddGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to stage files with pattern %s: %w", pattern, err)
	}

	return matchingFiles, nil
}

// Fallback: filtered staging for complex patterns
func (g *GitOperations) stageFiltered(
	worktree *git.Worktree,
	excludePatterns, includePatterns []string,
) ([]string, error) {
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Build list of files to stage (filtering phase)
	var filesToStage []string
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Worktree == git.Unmodified {
			continue
		}

		if shouldExcludeFile(file, excludePatterns) {
			continue
		}

		if len(includePatterns) > 0 && !shouldIncludeFile(file, includePatterns) {
			continue
		}

		filesToStage = append(filesToStage, file)
	}

	// Early return if no files to stage
	if len(filesToStage) == 0 {
		return []string{}, nil
	}

	// Stage files individually (necessary for complex filtering)
	for _, file := range filesToStage {
		_, err := worktree.Add(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stage file %s: %w", file, err)
		}
	}

	return filesToStage, nil
}

// Helper function to check if pattern is simple glob (no complex logic needed)
func isSimpleGlobPattern(pattern string) bool {
	// Simple check: if it contains only *, ?, and regular chars, it's probably a simple glob
	// Exclude patterns with path separators or complex logic
	return !strings.Contains(pattern, "/") &&
		(strings.Contains(pattern, "*") || strings.Contains(pattern, "?"))
}

func (g *GitOperations) GetStagedDiff() (string, error) {
	worktree, err := g.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	var diff strings.Builder
	for file := range status {
		fileStatus := status.File(file)
		if fileStatus.Staging == 0 {
			continue
		}
		diff.WriteString(fmt.Sprintf("--- a/%s\n+++ b/%s\n", file, file))
		diff.WriteString(fmt.Sprintf("@@ staged changes in %s @@\n", file))
	}

	return diff.String(), nil
}

func (g *GitOperations) CreateCommit(message string) error {
	// Get git configuration
	config, err := g.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get git config: %w", err)
	}

	worktree, err := g.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Create commit options with real user identity
	commitOptions := &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.UserName,
			Email: config.UserEmail,
			When:  time.Now(),
		},
	}

	// Add GPG signing if enabled
	if config.GPGSign {
		if config.SigningKey == "" {
			return fmt.Errorf("commit.gpgsign=true but user.signingkey not configured")
		}
		signKey, err := g.loadGPGSigningKey(config)
		if err != nil {
			return fmt.Errorf("failed to load GPG signing key %s: %w", config.SigningKey, err)
		}
		commitOptions.SignKey = signKey
	}

	_, err = worktree.Commit(message, commitOptions)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

// loadGPGSigningKey loads the GPG private key for commit signing
func (g *GitOperations) loadGPGSigningKey(config *GitConfig) (*openpgp.Entity, error) {
	if config.SigningKey == "" {
		return nil, fmt.Errorf("no signing key specified")
	}

	// Try to get GPG key from user's keyring
	keyring, err := g.getGPGKeyring(config.GPGProgram)
	if err != nil {
		return nil, fmt.Errorf("failed to access GPG keyring: %w", err)
	}

	// Find the specific signing key
	for _, entity := range keyring {
		// Check if this entity matches the signing key ID
		if g.matchesSigningKey(entity, config.SigningKey) {
			// Check if the key is encrypted and needs to be decrypted
			if entity.PrivateKey != nil && entity.PrivateKey.Encrypted {
				err := g.decryptPrivateKey(entity, config.SigningKey)
				if err != nil {
					return nil, fmt.Errorf("failed to decrypt signing key %s: %w", config.SigningKey, err)
				}
			}
			return entity, nil
		}
	}

	return nil, fmt.Errorf("signing key %s not found in keyring", config.SigningKey)
}

// getGPGKeyring reads the user's GPG keyring
func (g *GitOperations) getGPGKeyring(gpgProgram string) (openpgp.EntityList, error) {
	// Get GPG home directory
	gpgHome := os.Getenv("GNUPGHOME")
	if gpgHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		gpgHome = filepath.Join(homeDir, ".gnupg")
	}

	// Try to read secret keyring
	secretKeyringPath := filepath.Join(gpgHome, "secring.gpg")
	if _, err := os.Stat(secretKeyringPath); err == nil {
		// Old GPG format
		file, err := os.Open(secretKeyringPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open secret keyring: %w", err)
		}
		defer file.Close()

		return openpgp.ReadKeyRing(file)
	}

	// For newer GPG versions, we need to export keys
	return g.exportGPGKeys(gpgProgram)
}

// exportGPGKeys exports GPG keys using the gpg command
func (g *GitOperations) exportGPGKeys(gpgProgram string) (openpgp.EntityList, error) {
	// Export secret keys in ASCII armor format
	cmd := exec.Command(gpgProgram, "--export-secret-keys", "--armor")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to export GPG keys: %w", err)
	}

	// Parse the exported keys
	return openpgp.ReadArmoredKeyRing(strings.NewReader(string(output)))
}

// matchesSigningKey checks if a GPG entity matches the signing key identifier
func (g *GitOperations) matchesSigningKey(entity *openpgp.Entity, signingKey string) bool {
	// Check primary key ID (full or short form)
	primaryKeyID := fmt.Sprintf("%016X", entity.PrimaryKey.KeyId)
	if strings.HasSuffix(primaryKeyID, strings.ToUpper(signingKey)) {
		return true
	}

	// Check subkey IDs
	for _, subkey := range entity.Subkeys {
		subkeyID := fmt.Sprintf("%016X", subkey.PublicKey.KeyId)
		if strings.HasSuffix(subkeyID, strings.ToUpper(signingKey)) {
			return true
		}
	}

	// Check user IDs (email addresses)
	for _, identity := range entity.Identities {
		if strings.Contains(identity.UserId.Email, signingKey) {
			return true
		}
	}

	return false
}

// decryptPrivateKey prompts for passphrase and decrypts the GPG private key
func (g *GitOperations) decryptPrivateKey(entity *openpgp.Entity, keyID string) error {
	fmt.Printf("Enter passphrase for GPG key %s: ", keyID)

	// Read passphrase securely (without echoing to terminal)
	passphrase, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read passphrase: %w", err)
	}
	fmt.Println() // Add newline after password input

	// Attempt to decrypt the private key with the passphrase
	err = entity.PrivateKey.Decrypt(passphrase)
	if err != nil {
		return fmt.Errorf("incorrect passphrase or decryption failed: %w", err)
	}

	// Also decrypt subkeys if they exist
	for _, subkey := range entity.Subkeys {
		if subkey.PrivateKey != nil && subkey.PrivateKey.Encrypted {
			err = subkey.PrivateKey.Decrypt(passphrase)
			if err != nil {
				return fmt.Errorf("failed to decrypt subkey: %w", err)
			}
		}
	}

	return nil
}

func shouldExcludeFile(file string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	basename := filepath.Base(file)
	for _, pattern := range patterns {
		// Fast string containment check first (most common case)
		if strings.Contains(file, pattern) || strings.Contains(basename, pattern) {
			return true
		}
		// Expensive glob matching only if simple checks fail
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}
	return false
}

func shouldIncludeFile(file string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	basename := filepath.Base(file)
	for _, pattern := range patterns {
		// Fast string containment check first (most common case)
		if strings.Contains(file, pattern) || strings.Contains(basename, pattern) {
			return true
		}
		// Expensive glob matching only if simple checks fail
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
	}
	return false
}
