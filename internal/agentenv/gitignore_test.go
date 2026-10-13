package agentenv

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/go42-dev/go42x/assets"
)

func TestGitignoreReplacesBundledBlock(t *testing.T) {
	block := assets.AgentEnvGitIgnore()
	for _, tt := range []struct {
		name, previous, want string
	}{
		{"missing file", "", block},
		{"unmarked file", "user-entry\n", "user-entry\n\n" + block},
		{"no trailing newline", "user-entry", "user-entry\n\n" + block},
		{"replace obsolete rules", "before\n# go42x -- START\nobsolete\n# go42x -- END\nafter\n",
			"before\n" + block + "after\n"},
		{"block at EOF", "# go42x -- START\nobsolete\n# go42x -- END", block},
		{"preserve outside rules", "AGENTS.md\n# go42x -- START\nobsolete\n# go42x -- END\n!AGENTS.md\n",
			"AGENTS.md\n" + block + "!AGENTS.md\n"},
		{"preserve CRLF outside block", "before\r\n# go42x -- START\r\nobsolete\r\n# go42x -- END\r\nafter\r\n",
			"before\r\n" + block + "after\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, ".gitignore")
			if tt.previous != "" {
				writeFile(t, path, tt.previous)
			}
			if err := updateGitIgnore(dir); err != nil {
				t.Fatal(err)
			}
			if got := string(readFile(t, path)); got != tt.want {
				t.Fatalf(".gitignore = %q, want %q", got, tt.want)
			}
			if err := updateGitIgnore(dir); err != nil {
				t.Fatal(err)
			}
			if got := string(readFile(t, path)); got != tt.want {
				t.Fatalf("repeated replacement changed .gitignore: %q", got)
			}
		})
	}
}

func TestGitignoreRejectsMalformedMarkers(t *testing.T) {
	for _, previous := range []string{
		"before\n# go42x -- START\nuser rules\n",
		"before\n# go42x -- END\nafter\n",
		"# go42x -- END\n# go42x -- START\n",
		"# go42x -- START\n# go42x -- START\n# go42x -- END\n",
		"# go42x -- START\n# go42x -- END\n# go42x -- END\n",
		assets.AgentEnvGitIgnore() + assets.AgentEnvGitIgnore(),
	} {
		t.Run(previous, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, ".gitignore")
			writeFile(t, path, previous)
			if err := updateGitIgnore(dir); err == nil {
				t.Fatal("malformed markers accepted")
			}
			if !bytes.Equal(readFile(t, path), []byte(previous)) {
				t.Fatal("invalid block changed .gitignore")
			}
			if _, err := os.Stat(filepath.Join(dir, ".go42x")); !os.IsNotExist(err) {
				t.Fatalf("malformed markers created backup state: %v", err)
			}
		})
	}
}
