package rules

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidateFile(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	agentsPath := filepath.Join(filepath.Dir(filename), "..", "..", "..", "AGENTS.md")

	v := NewValidator()
	result := v.ValidateFile(agentsPath)

	if !result.Valid {
		t.Errorf("AGENTS.md should be valid, got %d errors", len(result.Errors))
		for _, e := range result.Errors {
			if e.Severity == "error" {
				t.Logf("  [%s] %s: %s", e.Severity, e.Rule, e.Message)
			}
		}
	}

	errors, warnings, infos := result.Summary()
	t.Logf("Validation: %d errors, %d warnings, %d info", errors, warnings, infos)
}

func TestValidateRequiredRules(t *testing.T) {
	doc := &Document{
		Rules: []Rule{
			{ID: "rule-1", Number: 1},
			{ID: "rule-2", Number: 2},
			// Missing rules 3-6
		},
	}

	v := NewValidator()
	result := v.Validate(doc)

	if result.Valid {
		t.Error("should be invalid due to missing rules")
	}

	// Should have errors for missing rules 3, 4, 5, 6
	errorCount := 0
	for _, e := range result.Errors {
		if e.Severity == "error" {
			errorCount++
		}
	}

	// At least 4 missing rules + 1 missing section
	if errorCount < 4 {
		t.Errorf("expected at least 4 errors for missing rules, got %d", errorCount)
	}
}

func TestValidateRequiredSections(t *testing.T) {
	doc := &Document{
		Rules: []Rule{
			{ID: "rule-1", Number: 1},
			{ID: "rule-2", Number: 2},
			{ID: "rule-3", Number: 3},
			{ID: "rule-4", Number: 4},
			{ID: "rule-5", Number: 5},
			{ID: "rule-6", Number: 6},
		},
		Sections: []Section{
			// Missing "Landing the Plane"
		},
	}

	v := NewValidator()
	result := v.Validate(doc)

	if result.Valid {
		t.Error("should be invalid due to missing sections")
	}

	foundMissing := false
	for _, e := range result.Errors {
		if e.Severity == "error" && e.Rule == "Landing the Plane" {
			foundMissing = true
			break
		}
	}

	if !foundMissing {
		t.Error("should report 'Landing the Plane' as missing")
	}
}

func TestValidateCodeBlockLanguages(t *testing.T) {
	doc := &Document{
		Rules: []Rule{
			{
				ID:     "rule-1",
				Number: 1,
				Examples: []Example{
					{Language: "bash", Code: "echo hello"},
					{Language: "unknown-lang", Code: "something"},
				},
			},
			{ID: "rule-2", Number: 2},
			{ID: "rule-3", Number: 3},
			{ID: "rule-4", Number: 4},
			{ID: "rule-5", Number: 5},
			{ID: "rule-6", Number: 6},
			{ID: "landing-the-plane", Title: "Landing the Plane"},
		},
	}

	v := NewValidator()
	result := v.Validate(doc)

	foundWarning := false
	for _, e := range result.Errors {
		if e.Severity == "warning" && e.Rule == "rule-1" {
			foundWarning = true
			t.Logf("Found expected warning: %s", e.Message)
			break
		}
	}

	if !foundWarning {
		t.Error("should warn about unknown code block language")
	}
}

func TestValidatorWithFullDocument(t *testing.T) {
	doc := &Document{
		Title: "Agent Rules & Guidelines",
		Rules: []Rule{
			{ID: "rule-1", Number: 1, Title: "Permissions"},
			{
				ID:         "rule-2",
				Number:     2,
				Title:      "Issue Tracking with bd",
				Prohibited: []string{"NEVER use TodoWrite"},
				Examples:   []Example{{Language: "bash", Code: "bd ready"}},
			},
			{
				ID:         "rule-3",
				Number:     3,
				Title:      "Git Branch Strategy",
				Prohibited: []string{"NEVER commit directly to main"},
			},
			{ID: "rule-4", Number: 4, Title: "User Review Before Execution"},
			{ID: "rule-5", Number: 5, Title: "Context Usage Reporting"},
			{ID: "rule-6", Number: 6, Title: "Git Commit on Every Change"},
			{ID: "landing-the-plane", Number: 0, Title: "Landing the Plane"},
		},
	}

	v := NewValidator()
	result := v.Validate(doc)

	if !result.Valid {
		t.Error("complete document should be valid")
		for _, e := range result.Errors {
			t.Logf("  [%s] %s: %s", e.Severity, e.Rule, e.Message)
		}
	}
}

func TestValidateSummary(t *testing.T) {
	result := &ValidationResult{
		Errors: []ValidationError{
			{Severity: "error", Message: "error 1"},
			{Severity: "error", Message: "error 2"},
			{Severity: "warning", Message: "warning 1"},
			{Severity: "info", Message: "info 1"},
			{Severity: "info", Message: "info 2"},
			{Severity: "info", Message: "info 3"},
		},
	}

	errors, warnings, infos := result.Summary()

	if errors != 2 {
		t.Errorf("expected 2 errors, got %d", errors)
	}
	if warnings != 1 {
		t.Errorf("expected 1 warning, got %d", warnings)
	}
	if infos != 3 {
		t.Errorf("expected 3 infos, got %d", infos)
	}
}

func TestStringsOverlap(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"use todowrite for tracking", "never use todowrite", true},
		{"commit to main branch", "never push to main", true},
		{"use git", "use hg", false},
		{"short", "words", false},
	}

	for _, tt := range tests {
		got := stringsOverlap(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("stringsOverlap(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
