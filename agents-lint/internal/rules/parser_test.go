package rules

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Get AGENTS.md path relative to this test
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	agentsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "AGENTS.md")

	doc, err := ParseFile(agentsPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Check document title
	if doc.Title != "Agent Rules & Guidelines" {
		t.Errorf("expected title 'Agent Rules & Guidelines', got %q", doc.Title)
	}

	// Should have multiple rules
	if len(doc.Rules) < 5 {
		t.Errorf("expected at least 5 rules, got %d", len(doc.Rules))
	}

	t.Logf("Parsed %d rules and %d sections", len(doc.Rules), len(doc.Sections))

	// Log rule details for debugging
	for _, r := range doc.Rules {
		t.Logf("Rule: ID=%s Number=%d Title=%q Required=%d Prohibited=%d Examples=%d",
			r.ID, r.Number, r.Title, len(r.Required), len(r.Prohibited), len(r.Examples))
	}
}

func TestParseRule2(t *testing.T) {
	lines := []string{
		"# Agent Rules",
		"",
		"## Rule 2: Issue Tracking with bd",
		"**Use bd for ALL task tracking. NEVER use TodoWrite or TODO comments.**",
		"",
		"```bash",
		"bd ready                              # Show unblocked work",
		"bd create \"Title\" -t task -p 1 -d \"Description\"  # Create issue",
		"```",
		"",
		"Always commit `.beads/issues.jsonl` with code changes.",
	}

	doc, err := Parse(lines)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	rule := doc.GetRuleByNumber(2)
	if rule == nil {
		t.Fatal("Rule 2 not found")
	}

	if rule.ID != "rule-2" {
		t.Errorf("expected ID 'rule-2', got %q", rule.ID)
	}

	if rule.Title != "Issue Tracking with bd" {
		t.Errorf("expected title 'Issue Tracking with bd', got %q", rule.Title)
	}

	// Should have prohibited behaviors (NEVER use TodoWrite)
	if len(rule.Prohibited) == 0 {
		t.Error("expected prohibited behaviors, got none")
	}

	// Should have examples
	if len(rule.Examples) == 0 {
		t.Error("expected examples, got none")
	} else if rule.Examples[0].Language != "bash" {
		t.Errorf("expected bash example, got %q", rule.Examples[0].Language)
	}
}

func TestParseCommitFormat(t *testing.T) {
	lines := []string{
		"# Agent Rules",
		"",
		"## Commit Message Format",
		"**Single-line commits only.** Git hook adds JIRA ID from branch name automatically.",
		"You MUST use single-line format.",
		"```bash",
		"git commit -m \"Add authentication middleware\"  # ✅ Correct",
		"```",
	}

	doc, err := Parse(lines)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should be parsed as a rule (has behavior indicators with MUST)
	rule := doc.GetRuleByID("commit-message-format")
	if rule == nil {
		// Check if it's in sections instead
		foundInSections := false
		for _, sec := range doc.Sections {
			if sec.Title == "Commit Message Format" {
				foundInSections = true
				t.Log("Found Commit Message Format in sections (no MUST/NEVER/ALWAYS)")
				break
			}
		}
		if !foundInSections {
			t.Fatal("commit-message-format not found in rules or sections")
		}
		return
	}

	if rule.Number != 0 {
		t.Errorf("expected Number 0 for non-numbered rule, got %d", rule.Number)
	}

	// Should have example marked as correct
	if len(rule.Examples) == 0 {
		t.Fatal("expected examples")
	}

	if !rule.Examples[0].IsCorrect {
		t.Error("expected example to be marked as correct")
	}
}

func TestExtractBehaviors(t *testing.T) {
	tests := []struct {
		name           string
		lines          []string
		wantRequired   int
		wantProhibited int
	}{
		{
			name: "NEVER pattern",
			lines: []string{
				"**NEVER commit directly to main.**",
			},
			wantRequired:   0,
			wantProhibited: 1,
		},
		{
			name: "MUST pattern",
			lines: []string{
				"You MUST request approval before starting.",
			},
			wantRequired:   1,
			wantProhibited: 0,
		},
		{
			name: "ALWAYS pattern",
			lines: []string{
				"ALWAYS commit after making changes.",
			},
			wantRequired:   1,
			wantProhibited: 0,
		},
		{
			name: "code block should be skipped",
			lines: []string{
				"Some text",
				"```bash",
				"NEVER do this in code example",
				"```",
			},
			wantRequired:   0,
			wantProhibited: 0,
		},
		{
			name: "mixed behaviors",
			lines: []string{
				"**Use bd for ALL task tracking. NEVER use TodoWrite.**",
				"ALWAYS commit changes.",
			},
			wantRequired:   1,
			wantProhibited: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, pro := extractBehaviors(tt.lines)
			if len(req) != tt.wantRequired {
				t.Errorf("required: got %d, want %d", len(req), tt.wantRequired)
			}
			if len(pro) != tt.wantProhibited {
				t.Errorf("prohibited: got %d, want %d", len(pro), tt.wantProhibited)
			}
		})
	}
}

func TestExtractExamples(t *testing.T) {
	lines := []string{
		"Some description",
		"```bash",
		"git commit -m \"Message\"  # ✅ Correct",
		"```",
		"",
		"```",
		"git commit -m \"Multi",
		"line\"  # ❌ Wrong",
		"```",
	}

	examples := extractExamples(lines)
	if len(examples) != 2 {
		t.Fatalf("expected 2 examples, got %d", len(examples))
	}

	if examples[0].Language != "bash" {
		t.Errorf("first example: expected language 'bash', got %q", examples[0].Language)
	}
	if !examples[0].IsCorrect {
		t.Error("first example should be marked as correct")
	}

	if examples[1].Language != "" {
		t.Errorf("second example: expected empty language, got %q", examples[1].Language)
	}
	if !examples[1].IsIncorrect {
		t.Error("second example should be marked as incorrect")
	}
}

func TestExtractSteps(t *testing.T) {
	lines := []string{
		"When user says \"land the plane\":",
		"1. File beads for remaining work",
		"2. Run quality gates (tests, lint) if code changed",
		"3. Close finished issues",
		"Some other text",
		"4. Commit and push beads changes",
	}

	steps := extractSteps(lines)
	if len(steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(steps))
	}

	if steps[0] != "File beads for remaining work" {
		t.Errorf("step 1: got %q", steps[0])
	}
}

func TestNormalizeID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Issue Tracking with bd", "issue-tracking-with-bd"},
		{"Commit Message Format", "commit-message-format"},
		{"Rule 2: Something", "rule-2-something"},
		{"BEFORE ANYTHING ELSE", "before-anything-else"},
		{"Pass the Baton", "pass-the-baton"},
	}

	for _, tt := range tests {
		got := normalizeID(tt.input)
		if got != tt.want {
			t.Errorf("normalizeID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGetRuleByID(t *testing.T) {
	doc := &Document{
		Rules: []Rule{
			{ID: "rule-1", Number: 1, Title: "First"},
			{ID: "rule-2", Number: 2, Title: "Second"},
			{ID: "commit-format", Number: 0, Title: "Commit Format"},
		},
	}

	rule := doc.GetRuleByID("rule-2")
	if rule == nil {
		t.Fatal("rule-2 not found")
	}
	if rule.Title != "Second" {
		t.Errorf("expected title 'Second', got %q", rule.Title)
	}

	rule = doc.GetRuleByID("nonexistent")
	if rule != nil {
		t.Error("expected nil for nonexistent rule")
	}
}

func TestGetRuleByNumber(t *testing.T) {
	doc := &Document{
		Rules: []Rule{
			{ID: "rule-1", Number: 1, Title: "First"},
			{ID: "rule-2", Number: 2, Title: "Second"},
		},
	}

	rule := doc.GetRuleByNumber(2)
	if rule == nil {
		t.Fatal("rule 2 not found")
	}
	if rule.ID != "rule-2" {
		t.Errorf("expected ID 'rule-2', got %q", rule.ID)
	}

	rule = doc.GetRuleByNumber(99)
	if rule != nil {
		t.Error("expected nil for nonexistent rule number")
	}
}

func TestParseRealAGENTSmd(t *testing.T) {
	// Get AGENTS.md path
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	agentsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "AGENTS.md")

	doc, err := ParseFile(agentsPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	// Verify specific rules exist
	expectedRules := []struct {
		id       string
		number   int
		hasTitle bool
	}{
		{"rule-2", 2, true},
		{"rule-3", 3, true},
		{"rule-4", 4, true},
		{"rule-5", 5, true},
		{"rule-6", 6, true},
	}

	for _, exp := range expectedRules {
		rule := doc.GetRuleByNumber(exp.number)
		if rule == nil {
			t.Errorf("Rule %d not found", exp.number)
			continue
		}
		if rule.ID != exp.id {
			t.Errorf("Rule %d: expected ID %q, got %q", exp.number, exp.id, rule.ID)
		}
		if exp.hasTitle && rule.Title == "" {
			t.Errorf("Rule %d: expected title, got empty", exp.number)
		}
	}

	// Rule 2 should have prohibited behaviors (NEVER use TodoWrite)
	rule2 := doc.GetRuleByNumber(2)
	if rule2 != nil {
		foundTodoWriteProhibition := false
		for _, p := range rule2.Prohibited {
			if strings.Contains(strings.ToLower(p), "todowrite") {
				foundTodoWriteProhibition = true
				break
			}
		}
		if !foundTodoWriteProhibition {
			t.Error("Rule 2 should prohibit TodoWrite usage")
		}
	}

	// Rule 3 should have prohibited behaviors (NEVER commit to main)
	rule3 := doc.GetRuleByNumber(3)
	if rule3 != nil {
		foundMainProhibition := false
		for _, p := range rule3.Prohibited {
			if strings.Contains(strings.ToLower(p), "main") {
				foundMainProhibition = true
				break
			}
		}
		if !foundMainProhibition {
			t.Error("Rule 3 should prohibit commits to main")
		}
	}
}
