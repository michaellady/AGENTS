package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestSingleLineCommit_ID(t *testing.T) {
	c := &SingleLineCommit{}
	if c.ID() != "single-line-commit" {
		t.Errorf("ID() = %q, want %q", c.ID(), "single-line-commit")
	}
}

func TestSingleLineCommit_ValidCommits(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"simple message", `git commit -m "Add feature"`},
		{"with quotes", `git commit -m "Fix \"bug\" in code"`},
		{"single quotes", `git commit -m 'Add feature'`},
		{"with flags", `git add . && git commit -m "Update" && git push`},
	}

	c := &SingleLineCommit{}

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

func TestSingleLineCommit_HeredocViolation(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{
			"cat heredoc",
			`git commit -m "$(cat <<'EOF'
Multi-line
message
EOF
)"`,
		},
		{
			"simple heredoc",
			`git commit -m "$(cat <<EOF
message
EOF
)"`,
		},
	}

	c := &SingleLineCommit{}

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
				t.Errorf("expected 1 violation, got %d", len(violations))
				return
			}
			if violations[0].Rule != "Commit Message Format" {
				t.Errorf("Rule = %q, want %q", violations[0].Rule, "Commit Message Format")
			}
		})
	}
}

func TestSingleLineCommit_MultilineViolation(t *testing.T) {
	cmd := "git commit -m \"Line 1\nLine 2\""
	input, _ := json.Marshal(BashInput{Command: cmd})
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Bash", Input: input},
		},
	}

	c := &SingleLineCommit{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
		return
	}
	if violations[0].Severity != SeverityError {
		t.Errorf("Severity = %v, want Error", violations[0].Severity)
	}
}

func TestSingleLineCommit_NonGitCommands(t *testing.T) {
	commands := []string{
		"git status",
		"git add .",
		"git push",
		"echo commit",
	}

	c := &SingleLineCommit{}

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

func TestSingleLineCommit_NonBashTools(t *testing.T) {
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Read", Input: json.RawMessage(`{"file_path":"/test"}`)},
			{ID: "t2", Name: "Write", Input: json.RawMessage(`{"file_path":"/test"}`)},
		},
	}

	c := &SingleLineCommit{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for non-Bash tools, got %d", len(violations))
	}
}

func TestSingleLineCommit_Registered(t *testing.T) {
	c := GetByID("single-line-commit")
	if c == nil {
		t.Error("single-line-commit checker not registered")
	}
}
