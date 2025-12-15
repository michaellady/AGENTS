package checker

import (
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func TestClarifyingQuestions_ID(t *testing.T) {
	c := &ClarifyingQuestions{}
	if c.ID() != "clarifying-questions" {
		t.Errorf("ID() = %q, want %q", c.ID(), "clarifying-questions")
	}
}

func TestClarifyingQuestions_WithQuestions(t *testing.T) {
	tests := []struct {
		name        string
		questionMsg string
	}{
		{"before I start", "Before I start implementing, I have a few questions about the approach."},
		{"which approach", "Which approach would you prefer for the authentication system?"},
		{"would you like", "Would you like me to use JWT tokens or session cookies?"},
		{"should I", "Should I add rate limiting to the login endpoint?"},
		{"option a/b", "Option A: JWT tokens. Option B: Session cookies. Which do you prefer?"},
	}

	c := &ClarifyingQuestions{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &transcript.Transcript{
				Events: []any{
					transcript.UserEvent{
						Event: transcript.Event{UUID: "e1"},
						Message: transcript.UserMessage{
							Content: []transcript.UserContentBlock{
								{Type: "text", Text: "Add user authentication to this app."},
							},
						},
					},
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e2"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "text", Text: tt.questionMsg},
							},
						},
					},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 0 {
				t.Errorf("expected 0 violations when questions asked %q, got %d", tt.questionMsg, len(violations))
			}
		})
	}
}

func TestClarifyingQuestions_WithAskUserQuestionTool(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I need to clarify a few things."},
						{Type: "tool_use", ID: "t1", Name: "AskUserQuestion"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when AskUserQuestion used, got %d", len(violations))
	}
}

func TestClarifyingQuestions_WithoutQuestions(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Let me start implementing the authentication system."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 13" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 13")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
}

func TestClarifyingQuestions_SimpleTask(t *testing.T) {
	// Simple tasks that don't match complex patterns shouldn't require questions
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Fix the typo in README.md"},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Let me start fixing that typo."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for simple task, got %d", len(violations))
	}
}

func TestClarifyingQuestions_ComplexTaskPatterns(t *testing.T) {
	complexTasks := []string{
		"Add user authentication",
		"Implement the payment system feature",
		"Refactor the database layer",
		"Optimize the search functionality",
		"Improve the performance of the API",
		"Add a new endpoint for user management",
		"Build a new notification module",
		"Create a new caching service",
		"Integrate with the payment gateway",
		"Migrate the database to PostgreSQL",
	}

	c := &ClarifyingQuestions{}

	for _, task := range complexTasks {
		t.Run(task, func(t *testing.T) {
			tr := &transcript.Transcript{
				Events: []any{
					transcript.UserEvent{
						Event: transcript.Event{UUID: "e1"},
						Message: transcript.UserMessage{
							Content: []transcript.UserContentBlock{
								{Type: "text", Text: task},
							},
						},
					},
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e2"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "text", Text: "I'll start implementing this right away."},
							},
						},
					},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 1 {
				t.Errorf("expected 1 violation for complex task %q, got %d", task, len(violations))
			}
		})
	}
}

func TestClarifyingQuestions_EmptyTranscript(t *testing.T) {
	tr := &transcript.Transcript{Events: []any{}}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for empty transcript, got %d", len(violations))
	}
}

func TestClarifyingQuestions_Registered(t *testing.T) {
	c := GetByID("clarifying-questions")
	if c == nil {
		t.Error("clarifying-questions checker not registered")
	}
}

// Edge case: Agent explores codebase (Read tool) then implements without questions
func TestClarifyingQuestions_ExplorationThenImplementation(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			// Agent explores first (no question, no implementation)
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Let me examine the codebase structure."},
						{Type: "tool_use", ID: "t1", Name: "Read"},
					},
				},
			},
			// Then implements without asking questions
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e3"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll start implementing the authentication system."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for exploration then implementation, got %d", len(violations))
	}
}

// Edge case: Questions and implementation in the same message should NOT be a violation
func TestClarifyingQuestions_QuestionAndImplementSameMessage(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Should I use JWT or session cookies? Let me start implementing with JWT."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	// Question was asked (even in same message), so no violation
	if len(violations) != 0 {
		t.Errorf("expected 0 violations when question asked in same message as implementation, got %d", len(violations))
	}
}

// Edge case: Question in first message, implementation in second message - should NOT be violation
func TestClarifyingQuestions_QuestionThenImplementLater(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Which authentication method would you prefer?"},
					},
				},
			},
			// User responds
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e3"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Use JWT please."},
					},
				},
			},
			// Now agent implements
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e4"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll start implementing JWT authentication."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when question was asked before implementation, got %d", len(violations))
	}
}

// Edge case: Write tool usage without implementation text should trigger violation
func TestClarifyingQuestions_WriteToolWithoutQuestions(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll add the authentication module."},
						{Type: "tool_use", ID: "t1", Name: "Write"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation when Write tool used without questions, got %d", len(violations))
	}
}

// Edge case: Edit tool usage without questions should trigger violation
func TestClarifyingQuestions_EditToolWithoutQuestions(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Refactor the database layer."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I see the issue."},
						{Type: "tool_use", ID: "t1", Name: "Edit"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation when Edit tool used without questions, got %d", len(violations))
	}
}

// Edge case: NotebookEdit tool usage without questions should trigger violation
func TestClarifyingQuestions_NotebookEditToolWithoutQuestions(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add data analysis feature to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll add the analysis code."},
						{Type: "tool_use", ID: "t1", Name: "NotebookEdit"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation when NotebookEdit tool used without questions, got %d", len(violations))
	}
}

// Edge case: Multiple complex tasks - each should be checked independently
func TestClarifyingQuestions_MultipleComplexTasks(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			// First task - handled correctly with questions
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Which auth method should I use?"},
					},
				},
			},
			// Second task - violation (no questions)
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e3"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Now refactor the database layer."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e4"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I'll start refactoring the database."},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	// Only second task should have violation
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for second task only, got %d", len(violations))
	}
}

// Edge case: No implementation at all - just exploration - should NOT be violation
func TestClarifyingQuestions_ExplorationOnly(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication to this app."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "Let me examine the current codebase structure first."},
						{Type: "tool_use", ID: "t1", Name: "Read"},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e3"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "text", Text: "I found the following files that may be relevant."},
						{Type: "tool_use", ID: "t2", Name: "Grep"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for exploration only (no implementation), got %d", len(violations))
	}
}

// Test additional implementation pattern variations
func TestClarifyingQuestions_ImplementationPatternVariations(t *testing.T) {
	implementationPhrases := []string{
		"I will now implement the feature.",
		"I will create the authentication module.",
		"Starting the implementation now.",
		"Now I'll create the new module.",
		"Now I'm going to implement this.",
	}

	c := &ClarifyingQuestions{}

	for _, phrase := range implementationPhrases {
		t.Run(phrase, func(t *testing.T) {
			tr := &transcript.Transcript{
				Events: []any{
					transcript.UserEvent{
						Event: transcript.Event{UUID: "e1"},
						Message: transcript.UserMessage{
							Content: []transcript.UserContentBlock{
								{Type: "text", Text: "Add user authentication."},
							},
						},
					},
					transcript.AssistantEvent{
						Event: transcript.Event{UUID: "e2"},
						Message: transcript.AssistantMessage{
							Content: []transcript.ContentBlock{
								{Type: "text", Text: phrase},
							},
						},
					},
				},
			}

			violations := c.Check(tr)
			if len(violations) != 1 {
				t.Errorf("expected 1 violation for implementation phrase %q, got %d", phrase, len(violations))
			}
		})
	}
}

// Test that questions with tool still work when combined with Write
func TestClarifyingQuestions_AskQuestionThenWrite(t *testing.T) {
	tr := &transcript.Transcript{
		Events: []any{
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e1"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Add user authentication."},
					},
				},
			},
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e2"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", ID: "t1", Name: "AskUserQuestion"},
					},
				},
			},
			// User responds
			transcript.UserEvent{
				Event: transcript.Event{UUID: "e3"},
				Message: transcript.UserMessage{
					Content: []transcript.UserContentBlock{
						{Type: "text", Text: "Use JWT."},
					},
				},
			},
			// Now writes
			transcript.AssistantEvent{
				Event: transcript.Event{UUID: "e4"},
				Message: transcript.AssistantMessage{
					Content: []transcript.ContentBlock{
						{Type: "tool_use", ID: "t2", Name: "Write"},
					},
				},
			},
		},
	}

	c := &ClarifyingQuestions{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations when AskUserQuestion used before Write, got %d", len(violations))
	}
}
