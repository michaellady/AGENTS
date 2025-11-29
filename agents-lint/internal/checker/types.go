// Package checker defines the interface and types for AGENTS.md rule checkers.
package checker

import "github.com/michaellady/agents-lint/internal/transcript"

// Severity indicates how serious a violation is.
type Severity int

const (
	// SeverityInfo is for informational findings that may not be violations.
	SeverityInfo Severity = iota
	// SeverityWarning is for violations that should be addressed but aren't critical.
	SeverityWarning
	// SeverityError is for clear rule violations that must be fixed.
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// Violation represents a single rule violation found in a transcript.
type Violation struct {
	// CheckerID is the unique identifier of the checker that found this violation.
	CheckerID string

	// Rule is the AGENTS.md rule number or name being violated.
	Rule string

	// Severity indicates how serious the violation is.
	Severity Severity

	// Message describes the violation in human-readable terms.
	Message string

	// EventUUID is the UUID of the event where the violation occurred.
	EventUUID string

	// ToolCallID is the tool_use ID if the violation is related to a tool call.
	ToolCallID string

	// Context provides additional details for debugging.
	Context map[string]string
}

// Checker is the interface that all rule checkers must implement.
type Checker interface {
	// ID returns a unique identifier for this checker (e.g., "no-todowrite").
	ID() string

	// Description returns a human-readable description of what this checker validates.
	Description() string

	// Check analyzes a transcript and returns any violations found.
	Check(t *transcript.Transcript) []Violation
}

// Result contains the output of running all checkers on a transcript.
type Result struct {
	// TranscriptPath is the path to the transcript file that was checked.
	TranscriptPath string

	// Violations is all violations found across all checkers.
	Violations []Violation

	// CheckersRun lists the IDs of all checkers that were executed.
	CheckersRun []string
}

// Summary returns counts of violations by severity.
func (r *Result) Summary() (errors, warnings, infos int) {
	for _, v := range r.Violations {
		switch v.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		case SeverityInfo:
			infos++
		}
	}
	return
}

// HasErrors returns true if any error-severity violations were found.
func (r *Result) HasErrors() bool {
	for _, v := range r.Violations {
		if v.Severity == SeverityError {
			return true
		}
	}
	return false
}
