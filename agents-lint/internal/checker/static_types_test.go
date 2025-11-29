package checker

import (
	"encoding/json"
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

// Helper to create json.RawMessage from a map
func toRawJSON(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

func TestStaticTypes_ID(t *testing.T) {
	c := &StaticTypes{}
	if c.ID() != "static-types" {
		t.Errorf("ID() = %q, want %q", c.ID(), "static-types")
	}
}

func TestStaticTypes_Description(t *testing.T) {
	c := &StaticTypes{}
	desc := c.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestStaticTypes_JSFileCreation(t *testing.T) {
	// Rule 9: Use TypeScript instead of JavaScript
	// Writing a .js file should trigger a warning
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/utils.js",
					"content":   "function hello() { return 'world'; }",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for .js file creation, got %d", len(violations))
	}

	v := violations[0]
	if v.Rule != "Rule 9" {
		t.Errorf("Rule = %q, want %q", v.Rule, "Rule 9")
	}
	if v.Severity != SeverityWarning {
		t.Errorf("Severity = %v, want Warning", v.Severity)
	}
	if v.ToolCallID != "tool-1" {
		t.Errorf("ToolCallID = %q, want %q", v.ToolCallID, "tool-1")
	}
}

func TestStaticTypes_TSFileAllowed(t *testing.T) {
	// TypeScript files should NOT trigger violations
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/utils.ts",
					"content":   "function hello(): string { return 'world'; }",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for .ts file, got %d", len(violations))
	}
}

func TestStaticTypes_JSXFileCreation(t *testing.T) {
	// .jsx should also trigger (prefer .tsx)
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/App.jsx",
					"content":   "export default function App() { return <div>Hello</div>; }",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for .jsx file creation, got %d", len(violations))
	}
}

func TestStaticTypes_TSXFileAllowed(t *testing.T) {
	// .tsx should NOT trigger
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/App.tsx",
					"content":   "export default function App(): JSX.Element { return <div>Hello</div>; }",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for .tsx file, got %d", len(violations))
	}
}

func TestStaticTypes_EditExistingJSAllowed(t *testing.T) {
	// Editing an EXISTING .js file is OK (legacy code maintenance)
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Edit",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path":  "/project/src/legacy.js",
					"old_string": "foo",
					"new_string": "bar",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for editing existing .js file, got %d", len(violations))
	}
}

func TestStaticTypes_GoFileAllowed(t *testing.T) {
	// Go files are preferred static types
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/cmd/main.go",
					"content":   "package main\n\nfunc main() {}",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for .go file, got %d", len(violations))
	}
}

func TestStaticTypes_RustFileAllowed(t *testing.T) {
	// Rust files are preferred static types
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/main.rs",
					"content":   "fn main() {}",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for .rs file, got %d", len(violations))
	}
}

func TestStaticTypes_KotlinFileAllowed(t *testing.T) {
	// Kotlin files are preferred static types
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/Main.kt",
					"content":   "fun main() {}",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for .kt file, got %d", len(violations))
	}
}

func TestStaticTypes_ConfigJSAllowed(t *testing.T) {
	// Config files like webpack.config.js, jest.config.js are exceptions
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/webpack.config.js",
					"content":   "module.exports = {}",
				}),
			},
			{
				ID:        "tool-2",
				Name:      "Write",
				EventUUID: "e2",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/jest.config.js",
					"content":   "module.exports = {}",
				}),
			},
			{
				ID:        "tool-3",
				Name:      "Write",
				EventUUID: "e3",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/.eslintrc.js",
					"content":   "module.exports = {}",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 0 {
		t.Errorf("expected 0 violations for config .js files, got %d", len(violations))
	}
}

func TestStaticTypes_MultipleViolations(t *testing.T) {
	// Multiple JS files should each trigger a violation
	tr := &transcript.Transcript{
		ToolCalls: []transcript.ToolCall{
			{
				ID:        "tool-1",
				Name:      "Write",
				EventUUID: "e1",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/a.js",
					"content":   "const a = 1;",
				}),
			},
			{
				ID:        "tool-2",
				Name:      "Write",
				EventUUID: "e2",
				Input: toRawJSON(map[string]any{
					"file_path": "/project/src/b.js",
					"content":   "const b = 2;",
				}),
			},
		},
	}

	c := &StaticTypes{}
	violations := c.Check(tr)

	if len(violations) != 2 {
		t.Errorf("expected 2 violations for two .js files, got %d", len(violations))
	}
}

func TestStaticTypes_Registered(t *testing.T) {
	c := GetByID("static-types")
	if c == nil {
		t.Error("static-types checker not registered")
	}
}
