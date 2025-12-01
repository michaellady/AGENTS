package rules

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a single validation issue.
type ValidationError struct {
	// Severity indicates how serious this issue is.
	Severity string // "error", "warning", "info"

	// Rule is the rule ID or section name where the issue was found.
	Rule string

	// Message describes the validation issue.
	Message string

	// Line is the approximate line number (0 if unknown).
	Line int
}

// ValidationResult contains the output of validating an AGENTS.md file.
type ValidationResult struct {
	// Valid is true if no errors were found (warnings are allowed).
	Valid bool

	// Errors contains all validation issues found.
	Errors []ValidationError

	// Document is the parsed document (nil if parse failed).
	Document *Document
}

// Summary returns counts of issues by severity.
func (r *ValidationResult) Summary() (errors, warnings, infos int) {
	for _, e := range r.Errors {
		switch e.Severity {
		case "error":
			errors++
		case "warning":
			warnings++
		case "info":
			infos++
		}
	}
	return
}

// Validator validates AGENTS.md files against structural requirements.
type Validator struct {
	// RequiredRules lists rule numbers that must be present.
	RequiredRules []int

	// RequiredSections lists section titles that must be present.
	RequiredSections []string

	// ValidLanguages lists recognized code block languages.
	ValidLanguages map[string]bool

	// KnownCommands lists commands that should be referenced consistently.
	KnownCommands map[string]bool
}

// NewValidator creates a validator with default configuration.
func NewValidator() *Validator {
	return &Validator{
		RequiredRules: []int{1, 2, 3, 4, 5, 6},
		RequiredSections: []string{
			"Landing the Plane",
		},
		ValidLanguages: map[string]bool{
			"":           true, // no language specified
			"bash":       true,
			"sh":         true,
			"go":         true,
			"python":     true,
			"javascript": true,
			"typescript": true,
			"json":       true,
			"yaml":       true,
			"markdown":   true,
			"sql":        true,
			"rust":       true,
			"kotlin":     true,
		},
		KnownCommands: map[string]bool{
			"bd":    true,
			"git":   true,
			"gh":    true,
			"go":    true,
			"npm":   true,
			"cargo": true,
		},
	}
}

// ValidateFile validates an AGENTS.md file at the given path.
func (v *Validator) ValidateFile(path string) *ValidationResult {
	doc, err := ParseFile(path)
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ValidationError{
				{Severity: "error", Message: fmt.Sprintf("Failed to parse file: %v", err)},
			},
		}
	}

	return v.Validate(doc)
}

// Validate checks a parsed document for structural issues.
func (v *Validator) Validate(doc *Document) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Document: doc,
	}

	// Check required rules exist
	v.checkRequiredRules(doc, result)

	// Check required sections exist
	v.checkRequiredSections(doc, result)

	// Validate code blocks
	v.checkCodeBlocks(doc, result)

	// Check for conflicting rules
	v.checkConflicts(doc, result)

	// Check command consistency
	v.checkCommands(doc, result)

	// Determine overall validity
	for _, e := range result.Errors {
		if e.Severity == "error" {
			result.Valid = false
			break
		}
	}

	return result
}

// checkRequiredRules verifies all required rules are present.
func (v *Validator) checkRequiredRules(doc *Document, result *ValidationResult) {
	for _, num := range v.RequiredRules {
		rule := doc.GetRuleByNumber(num)
		if rule == nil {
			result.Errors = append(result.Errors, ValidationError{
				Severity: "error",
				Rule:     fmt.Sprintf("Rule %d", num),
				Message:  fmt.Sprintf("Required Rule %d is missing", num),
			})
		}
	}
}

// checkRequiredSections verifies all required sections are present.
func (v *Validator) checkRequiredSections(doc *Document, result *ValidationResult) {
	for _, title := range v.RequiredSections {
		found := false
		for _, sec := range doc.Sections {
			if sec.Title == title {
				found = true
				break
			}
		}
		// Also check if it's parsed as a rule
		if !found {
			id := normalizeID(title)
			if doc.GetRuleByID(id) != nil {
				found = true
			}
		}
		if !found {
			result.Errors = append(result.Errors, ValidationError{
				Severity: "error",
				Rule:     title,
				Message:  fmt.Sprintf("Required section '%s' is missing", title),
			})
		}
	}
}

// checkCodeBlocks validates code block languages.
func (v *Validator) checkCodeBlocks(doc *Document, result *ValidationResult) {
	for _, rule := range doc.Rules {
		for _, ex := range rule.Examples {
			if !v.ValidLanguages[ex.Language] {
				result.Errors = append(result.Errors, ValidationError{
					Severity: "warning",
					Rule:     rule.ID,
					Message:  fmt.Sprintf("Unknown code block language: %q", ex.Language),
				})
			}
		}
	}
}

// checkConflicts looks for potentially conflicting rules.
func (v *Validator) checkConflicts(doc *Document, result *ValidationResult) {
	// Check for contradictory required/prohibited behaviors
	allRequired := make(map[string]string)  // behavior -> rule ID
	allProhibited := make(map[string]string) // behavior -> rule ID

	for _, rule := range doc.Rules {
		for _, req := range rule.Required {
			key := strings.ToLower(req)
			allRequired[key] = rule.ID
		}
		for _, pro := range rule.Prohibited {
			key := strings.ToLower(pro)
			allProhibited[key] = rule.ID
		}
	}

	// Look for overlapping terms in required and prohibited
	for req, reqRule := range allRequired {
		for pro, proRule := range allProhibited {
			// Check if they're about the same thing
			if stringsOverlap(req, pro) && reqRule != proRule {
				result.Errors = append(result.Errors, ValidationError{
					Severity: "warning",
					Rule:     reqRule,
					Message:  fmt.Sprintf("Potential conflict: %s requires something %s prohibits", reqRule, proRule),
				})
			}
		}
	}
}

// checkCommands looks for command consistency.
func (v *Validator) checkCommands(doc *Document, result *ValidationResult) {
	// Pattern to find command references in code blocks
	cmdPattern := regexp.MustCompile(`\b([a-z]+)\s+[a-z]`)

	commandUsage := make(map[string][]string) // command -> rule IDs

	for _, rule := range doc.Rules {
		for _, ex := range rule.Examples {
			matches := cmdPattern.FindAllStringSubmatch(ex.Code, -1)
			for _, m := range matches {
				cmd := m[1]
				if v.KnownCommands[cmd] {
					commandUsage[cmd] = append(commandUsage[cmd], rule.ID)
				}
			}
		}
	}

	// Check for bd command usage (should be in Rule 2)
	if rules, ok := commandUsage["bd"]; ok {
		hasRule2 := false
		for _, r := range rules {
			if r == "rule-2" {
				hasRule2 = true
				break
			}
		}
		if !hasRule2 {
			result.Errors = append(result.Errors, ValidationError{
				Severity: "info",
				Rule:     "rule-2",
				Message:  "bd command used in examples but Rule 2 doesn't have bd examples",
			})
		}
	}
}

// stringsOverlap checks if two strings share significant words.
func stringsOverlap(a, b string) bool {
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)

	for _, wa := range wordsA {
		if len(wa) < 4 {
			continue // skip short words
		}
		for _, wb := range wordsB {
			if strings.EqualFold(wa, wb) {
				return true
			}
		}
	}
	return false
}
