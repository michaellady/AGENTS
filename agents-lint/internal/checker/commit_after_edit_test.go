package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestCommitAfterEdit_ID(t *testing.T) {
	c := &CommitAfterEdit{}
	if c.ID() != "commit-after-edit" {
		t.Errorf("ID() = %q, want %q", c.ID(), "commit-after-edit")
	}
}

func TestCommitAfterEdit_EditThenCommit(t *testing.T) {
	editInput, _ := json.Marshal(map[string]string{"file_path": "/test/file.go"})
	commitInput, _ := json.Marshal(BashInput{Command: "git add -A && git commit -m \"Update\""})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Edit", Input: editInput, EventUUID: "e1"},
			{ID: "t2", Name: "Bash", Input: commitInput, EventUUID: "e2"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when edit followed by commit, got %d", len(violations))
	}
}

func TestCommitAfterEdit_MultipleEditsThenCommit(t *testing.T) {
	editInput, _ := json.Marshal(map[string]string{"file_path": "/test/file.go"})
	writeInput, _ := json.Marshal(map[string]string{"file_path": "/test/new.go"})
	commitInput, _ := json.Marshal(BashInput{Command: "git commit -am \"Update\""})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Edit", Input: editInput, EventUUID: "e1"},
			{ID: "t2", Name: "Write", Input: writeInput, EventUUID: "e2"},
			{ID: "t3", Name: "Bash", Input: commitInput, EventUUID: "e3"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when multiple edits followed by commit, got %d", len(violations))
	}
}

func TestCommitAfterEdit_EditWithoutCommit(t *testing.T) {
	editInput, _ := json.Marshal(map[string]string{"file_path": "/test/file.go"})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Edit", Input: editInput, EventUUID: "e1"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation when edit not committed, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 6" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 6")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
}

func TestCommitAfterEdit_WriteWithoutCommit(t *testing.T) {
	writeInput, _ := json.Marshal(map[string]string{"file_path": "/test/new.go"})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Write", Input: writeInput, EventUUID: "e1"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Errorf("expected 1 violation when write not committed, got %d", len(violations))
	}
}

func TestCommitAfterEdit_NotebookEditWithoutCommit(t *testing.T) {
	nbInput, _ := json.Marshal(map[string]string{"notebook_path": "/test/nb.ipynb"})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "NotebookEdit", Input: nbInput, EventUUID: "e1"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Errorf("expected 1 violation when notebook edit not committed, got %d", len(violations))
	}
}

func TestCommitAfterEdit_ReadDoesNotRequireCommit(t *testing.T) {
	readInput, _ := json.Marshal(map[string]string{"file_path": "/test/file.go"})

	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Read", Input: readInput, EventUUID: "e1"},
		},
	}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for Read tool, got %d", len(violations))
	}
}

func TestCommitAfterEdit_EmptyTranscript(t *testing.T) {
	tr := &transcript.Transcript{ToolCalls: []transcript.ToolCall{}}

	c := &CommitAfterEdit{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for empty transcript, got %d", len(violations))
	}
}

func TestCommitAfterEdit_Registered(t *testing.T) {
	c := GetByID("commit-after-edit")
	if c == nil {
		t.Error("commit-after-edit checker not registered")
	}
}
