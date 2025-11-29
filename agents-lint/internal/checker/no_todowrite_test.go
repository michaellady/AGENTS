package checker

import (
	"path/filepath"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestNoTodoWrite_ID(t *testing.T) {
	c := &NoTodoWrite{}
	if c.ID() != "no-todowrite" {
		t.Errorf("ID() = %q, want %q", c.ID(), "no-todowrite")
	}
}

func TestNoTodoWrite_Description(t *testing.T) {
	c := &NoTodoWrite{}
	if c.Description() == "" {
		t.Error("Description() is empty")
	}
}

func TestNoTodoWrite_PassingTranscript(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "transcripts", "passing", "simple.ndjson")
	tr, err := transcript.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	c := &NoTodoWrite{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(violations), violations)
	}
}

func TestNoTodoWrite_FailingTranscript(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "transcripts", "failing", "uses-todowrite.ndjson")
	tr, err := transcript.ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	c := &NoTodoWrite{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.CheckerID != "no-todowrite" {
		t.Errorf("CheckerID = %q, want %q", v.CheckerID, "no-todowrite")
	}
	if v.Rule != "Rule 2" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 2")
	}
	if v.Severity != SeverityError {
		t.Errorf("Severity = %v, want %v", v.Severity, SeverityError)
	}
	if v.ToolCallID != "tool-1" {
		t.Errorf("ToolCallID = %q, want %q", v.ToolCallID, "tool-1")
	}
}

func TestNoTodoWrite_MultipleCalls(t *testing.T) {
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{ID: "t1", Name: "Read"},
			{ID: "t2", Name: "TodoWrite", EventUUID: "e1"},
			{ID: "t3", Name: "Write"},
			{ID: "t4", Name: "TodoWrite", EventUUID: "e2"},
		},
	}

	c := &NoTodoWrite{}
	violations := c.Check(tr)

	if len(violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(violations))
	}
}

func TestNoTodoWrite_Registered(t *testing.T) {
	c := GetByID("no-todowrite")
	if c == nil {
		t.Error("no-todowrite checker not registered")
	}
}
