// Package rules parses AGENTS.md into structured rule definitions.
package rules

// Rule represents a parsed rule from AGENTS.md.
type Rule struct {
	// ID is a normalized identifier (e.g., "rule-2", "commit-format").
	ID string

	// Number is the rule number if applicable (0 for non-numbered rules).
	Number int

	// Title is the rule title (e.g., "Issue Tracking with bd").
	Title string

	// RawTitle is the original heading text (e.g., "Rule 2: Issue Tracking with bd").
	RawTitle string

	// Description is the full text content of the rule section.
	Description string

	// Required lists behaviors the agent MUST do.
	Required []string

	// Prohibited lists behaviors the agent must NOT do.
	Prohibited []string

	// Examples contains code examples from the rule.
	Examples []Example
}

// Example represents a code example from a rule section.
type Example struct {
	// Language is the code block language (e.g., "bash", "").
	Language string

	// Code is the example code content.
	Code string

	// IsCorrect indicates if this is a correct example (has checkmark).
	IsCorrect bool

	// IsIncorrect indicates if this is an incorrect example (has X mark).
	IsIncorrect bool
}

// Document represents the entire parsed AGENTS.md file.
type Document struct {
	// Title is the document title from the H1 heading.
	Title string

	// Rules contains all parsed rules in order.
	Rules []Rule

	// Sections contains non-rule sections (like "Landing the Plane").
	Sections []Section
}

// Section represents a non-rule section of the document.
type Section struct {
	// Title is the section heading.
	Title string

	// Content is the section body text.
	Content string

	// Steps contains numbered steps if present.
	Steps []string
}
