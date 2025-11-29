package checker

import (
	"testing"

	"github.com/michaellady/agents-lint/internal/transcript"
)

// mockChecker is a simple checker for testing.
type mockChecker struct {
	id          string
	description string
	violations  []Violation
}

func (m *mockChecker) ID() string          { return m.id }
func (m *mockChecker) Description() string { return m.description }
func (m *mockChecker) Check(_ *transcript.Transcript) []Violation {
	return m.violations
}

// saveRegistry saves and restores the registry for isolated testing.
func saveRegistry(t *testing.T) {
	snap := snapshot()
	t.Cleanup(func() { restore(snap) })
}

func TestRegisterAndGetByID(t *testing.T) {
	saveRegistry(t)
	Clear()

	c := &mockChecker{id: "test-checker", description: "A test checker"}
	Register(c)

	got := GetByID("test-checker")
	if got == nil {
		t.Fatal("GetByID returned nil")
	}
	if got.ID() != "test-checker" {
		t.Errorf("ID() = %q, want %q", got.ID(), "test-checker")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	saveRegistry(t)
	Clear()

	got := GetByID("nonexistent")
	if got != nil {
		t.Errorf("GetByID returned %v, want nil", got)
	}
}

func TestGetAll(t *testing.T) {
	saveRegistry(t)
	Clear()

	Register(&mockChecker{id: "checker-b"})
	Register(&mockChecker{id: "checker-a"})
	Register(&mockChecker{id: "checker-c"})

	checkers := GetAll()
	if len(checkers) != 3 {
		t.Fatalf("len(GetAll()) = %d, want 3", len(checkers))
	}

	// Should be sorted by ID
	ids := make([]string, len(checkers))
	for i, c := range checkers {
		ids[i] = c.ID()
	}
	expected := []string{"checker-a", "checker-b", "checker-c"}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("GetAll()[%d].ID() = %q, want %q", i, id, expected[i])
		}
	}
}

func TestIDs(t *testing.T) {
	saveRegistry(t)
	Clear()

	Register(&mockChecker{id: "z-checker"})
	Register(&mockChecker{id: "a-checker"})

	ids := IDs()
	if len(ids) != 2 {
		t.Fatalf("len(IDs()) = %d, want 2", len(ids))
	}
	if ids[0] != "a-checker" || ids[1] != "z-checker" {
		t.Errorf("IDs() = %v, want [a-checker z-checker]", ids)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	saveRegistry(t)
	Clear()

	Register(&mockChecker{id: "dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()

	Register(&mockChecker{id: "dup"})
}

func TestRunAll(t *testing.T) {
	saveRegistry(t)
	Clear()

	Register(&mockChecker{
		id: "checker-1",
		violations: []Violation{
			{CheckerID: "checker-1", Severity: SeverityError, Message: "error 1"},
		},
	})
	Register(&mockChecker{
		id: "checker-2",
		violations: []Violation{
			{CheckerID: "checker-2", Severity: SeverityWarning, Message: "warning 1"},
			{CheckerID: "checker-2", Severity: SeverityInfo, Message: "info 1"},
		},
	})

	result := RunAll(&transcript.Transcript{})

	if len(result.CheckersRun) != 2 {
		t.Errorf("len(CheckersRun) = %d, want 2", len(result.CheckersRun))
	}
	if len(result.Violations) != 3 {
		t.Errorf("len(Violations) = %d, want 3", len(result.Violations))
	}

	errors, warnings, infos := result.Summary()
	if errors != 1 || warnings != 1 || infos != 1 {
		t.Errorf("Summary() = (%d, %d, %d), want (1, 1, 1)", errors, warnings, infos)
	}

	if !result.HasErrors() {
		t.Error("HasErrors() = false, want true")
	}
}

func TestRunByIDs(t *testing.T) {
	saveRegistry(t)
	Clear()

	Register(&mockChecker{id: "run-me", violations: []Violation{{Message: "found"}}})
	Register(&mockChecker{id: "skip-me", violations: []Violation{{Message: "skipped"}}})

	result := RunByIDs(&transcript.Transcript{}, []string{"run-me", "nonexistent"})

	if len(result.CheckersRun) != 1 {
		t.Errorf("len(CheckersRun) = %d, want 1", len(result.CheckersRun))
	}
	if result.CheckersRun[0] != "run-me" {
		t.Errorf("CheckersRun[0] = %q, want %q", result.CheckersRun[0], "run-me")
	}
	if len(result.Violations) != 1 {
		t.Errorf("len(Violations) = %d, want 1", len(result.Violations))
	}
}

func TestResultHasErrors_NoErrors(t *testing.T) {
	result := &Result{
		Violations: []Violation{
			{Severity: SeverityWarning},
			{Severity: SeverityInfo},
		},
	}

	if result.HasErrors() {
		t.Error("HasErrors() = true, want false")
	}
}
