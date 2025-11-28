package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestGitBranch_ID(t *testing.T) {
	c := &GitBranch{}
	if c.ID() != "git-branch" {
		t.Errorf("ID() = %q, want %q", c.ID(), "git-branch")
	}
}

func TestGitBranch_ValidCommands(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"push to feature branch", "git push -u origin feature-123"},
		{"push to remote branch", "git push origin my-branch"},
		{"checkout feature", "git checkout -b feature-branch"},
		{"checkout existing", "git checkout my-feature"},
		{"push with upstream", "git push --set-upstream origin my-branch"},
		{"regular push", "git push"},
	}

	c := &GitBranch{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, _ := json.Marshal(BashInput{Command: tt.command})
			tr := &transcript.Transcript{
				ToolCalls: []transcript.ToolCall{
					{ID: "t1", Name: "Bash", Input: input},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 0 {
				t.Errorf("expected 0 violations for %q, got %d: %v", tt.command, len(violations), violations)
			}
		})
	}
}

func TestGitBranch_PushMainViolation(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"push origin main", "git push origin main"},
		{"push origin master", "git push origin master"},
		{"push with flags main", "git push -u origin main"},
	}

	c := &GitBranch{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, _ := json.Marshal(BashInput{Command: tt.command})
			tr := &transcript.Transcript{
				ToolCalls: []transcript.ToolCall{
					{ID: "t1", Name: "Bash", Input: input},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 1 {
				t.Errorf("expected 1 violation for %q, got %d", tt.command, len(violations))
				return
			}
			if violations[0].Rule != "Rule 3" {
				t.Errorf("Rule = %q, want %q", violations[0].Rule, "Rule 3")
			}
			if violations[0].Severity != SeverityError {
				t.Errorf("Severity = %v, want Error", violations[0].Severity)
			}
		})
	}
}

func TestGitBranch_ForcePushMainViolation(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"force push main", "git push -f origin main"},
		{"force push master", "git push --force origin master"},
		{"force with other flags", "git push -u -f origin main"},
	}

	c := &GitBranch{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, _ := json.Marshal(BashInput{Command: tt.command})
			tr := &transcript.Transcript{
				ToolCalls: []transcript.ToolCall{
					{ID: "t1", Name: "Bash", Input: input},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 1 {
				t.Errorf("expected 1 violation for %q, got %d", tt.command, len(violations))
				return
			}
			// Force push message should mention "dangerous"
			if violations[0].Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestGitBranch_NonGitCommands(t *testing.T) {
	commands := []string{
		"echo main",
		"cat main.go",
		"ls -la",
	}

	c := &GitBranch{}

	for _, cmd := range commands {
		input, _ := json.Marshal(BashInput{Command: cmd})
		tr := &transcript.Transcript{
			ToolCalls: []transcript.ToolCall{
				{ID: "t1", Name: "Bash", Input: input},
			},
		}

		violations := c.Check(tr)
		if len(violations) != 0 {
			t.Errorf("expected 0 violations for %q, got %d", cmd, len(violations))
		}
	}
}

func TestGitBranch_Registered(t *testing.T) {
	c := GetByID("git-branch")
	if c == nil {
		t.Error("git-branch checker not registered")
	}
}
