package rules

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Patterns for parsing AGENTS.md
var (
	// h1Pattern matches "# Title"
	h1Pattern = regexp.MustCompile(`^#\s+(.+)$`)

	// h2Pattern matches "## Title" or "## Rule N: Title"
	h2Pattern = regexp.MustCompile(`^##\s+(.+)$`)

	// rulePattern matches "Rule N: Title" in headings
	rulePattern = regexp.MustCompile(`^Rule\s+(\d+):\s+(.+)$`)

	// codeBlockStart matches ```language (with optional trailing content)
	codeBlockStart = regexp.MustCompile("^```(\\w*)")

	// codeBlockEnd matches ``` alone on a line
	codeBlockEnd = regexp.MustCompile("^```\\s*$")

	// boldPattern matches **text**
	boldPattern = regexp.MustCompile(`\*\*([^*]+)\*\*`)

	// neverPattern matches NEVER (case insensitive with word boundaries)
	neverPattern = regexp.MustCompile(`(?i)\bNEVER\b`)

	// mustPattern matches MUST or "must" in bold
	mustPattern = regexp.MustCompile(`(?i)\bMUST\b|\*\*[^*]*must[^*]*\*\*`)

	// alwaysPattern matches ALWAYS or "always" in bold
	alwaysPattern = regexp.MustCompile(`(?i)\bALWAYS\b|\*\*[^*]*always[^*]*\*\*`)

	// correctMarker matches checkmark emoji
	correctMarker = regexp.MustCompile(`✅|# ✅`)

	// incorrectMarker matches X emoji
	incorrectMarker = regexp.MustCompile(`❌|# ❌`)

	// numberedStepPattern matches "1. Step text"
	numberedStepPattern = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
)

// ParseFile reads and parses an AGENTS.md file.
func ParseFile(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return Parse(lines)
}

// Parse parses AGENTS.md content from lines.
func Parse(lines []string) (*Document, error) {
	doc := &Document{}

	// Find document title (H1)
	for _, line := range lines {
		if m := h1Pattern.FindStringSubmatch(line); m != nil {
			doc.Title = m[1]
			break
		}
	}

	// Split into sections by H2 headings
	sections := splitByH2(lines)

	for _, sec := range sections {
		if len(sec.lines) == 0 {
			continue
		}

		heading := sec.heading
		content := strings.Join(sec.lines, "\n")

		// Check if this is a numbered rule
		if m := rulePattern.FindStringSubmatch(heading); m != nil {
			num, _ := strconv.Atoi(m[1])
			rule := Rule{
				ID:          fmt.Sprintf("rule-%d", num),
				Number:      num,
				Title:       m[2],
				RawTitle:    heading,
				Description: content,
			}
			rule.Required, rule.Prohibited = extractBehaviors(sec.lines)
			rule.Examples = extractExamples(sec.lines)
			doc.Rules = append(doc.Rules, rule)
		} else {
			// Non-numbered section - could still be a rule-like section
			id := normalizeID(heading)

			// Check if it looks like a rule (has required/prohibited behaviors)
			required, prohibited := extractBehaviors(sec.lines)
			if len(required) > 0 || len(prohibited) > 0 {
				rule := Rule{
					ID:          id,
					Number:      0,
					Title:       heading,
					RawTitle:    heading,
					Description: content,
					Required:    required,
					Prohibited:  prohibited,
					Examples:    extractExamples(sec.lines),
				}
				doc.Rules = append(doc.Rules, rule)
			} else {
				// Regular section
				section := Section{
					Title:   heading,
					Content: content,
					Steps:   extractSteps(sec.lines),
				}
				doc.Sections = append(doc.Sections, section)
			}
		}
	}

	return doc, nil
}

// sectionData holds lines for a section.
type sectionData struct {
	heading string
	lines   []string
}

// splitByH2 splits lines into sections based on H2 headings.
func splitByH2(lines []string) []sectionData {
	var sections []sectionData
	var current *sectionData

	for _, line := range lines {
		if m := h2Pattern.FindStringSubmatch(line); m != nil {
			if current != nil {
				sections = append(sections, *current)
			}
			current = &sectionData{heading: m[1]}
		} else if current != nil {
			current.lines = append(current.lines, line)
		}
	}
	if current != nil {
		sections = append(sections, *current)
	}

	return sections
}

// extractBehaviors finds required and prohibited behaviors from lines.
func extractBehaviors(lines []string) (required, prohibited []string) {
	inCodeBlock := false

	for _, line := range lines {
		// Track code blocks to avoid extracting from examples
		if codeBlockStart.MatchString(line) {
			inCodeBlock = true
			continue
		}
		if codeBlockEnd.MatchString(line) {
			inCodeBlock = false
			continue
		}
		if inCodeBlock {
			continue
		}

		// Look for NEVER patterns
		if neverPattern.MatchString(line) {
			// Extract the prohibition
			behavior := extractBehaviorText(line, "NEVER")
			if behavior != "" {
				prohibited = append(prohibited, behavior)
			}
		}

		// Look for MUST/ALWAYS patterns
		if mustPattern.MatchString(line) || alwaysPattern.MatchString(line) {
			behavior := extractBehaviorText(line, "MUST")
			if behavior != "" {
				required = append(required, behavior)
			}
		}
	}

	return required, prohibited
}

// extractBehaviorText extracts the behavior description from a line.
func extractBehaviorText(line, keyword string) string {
	// Remove markdown formatting
	text := boldPattern.ReplaceAllString(line, "$1")
	text = strings.TrimSpace(text)

	// Skip empty or very short lines
	if len(text) < 10 {
		return ""
	}

	return text
}

// extractExamples finds code examples from lines.
func extractExamples(lines []string) []Example {
	var examples []Example
	var current *Example
	var codeLines []string

	for _, line := range lines {
		// Check for code block end first (if we're in a block)
		if current != nil && codeBlockEnd.MatchString(line) {
			current.Code = strings.Join(codeLines, "\n")

			// Check for correct/incorrect markers in the code
			if correctMarker.MatchString(current.Code) {
				current.IsCorrect = true
			}
			if incorrectMarker.MatchString(current.Code) {
				current.IsIncorrect = true
			}

			examples = append(examples, *current)
			current = nil
			codeLines = nil
			continue
		}

		// Check for code block start (if not already in a block)
		if current == nil {
			if m := codeBlockStart.FindStringSubmatch(line); m != nil {
				current = &Example{Language: m[1]}
				codeLines = nil
				continue
			}
		}

		// Collect code lines if in a block
		if current != nil {
			codeLines = append(codeLines, line)
		}
	}

	return examples
}

// extractSteps finds numbered steps from lines.
func extractSteps(lines []string) []string {
	var steps []string
	for _, line := range lines {
		if m := numberedStepPattern.FindStringSubmatch(line); m != nil {
			steps = append(steps, m[2])
		}
	}
	return steps
}

// normalizeID converts a title to a normalized ID.
func normalizeID(title string) string {
	// Convert to lowercase
	id := strings.ToLower(title)

	// Replace spaces and special chars with hyphens
	id = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(id, "-")

	// Trim leading/trailing hyphens
	id = strings.Trim(id, "-")

	return id
}

// GetRuleByID finds a rule by its ID.
func (d *Document) GetRuleByID(id string) *Rule {
	for i := range d.Rules {
		if d.Rules[i].ID == id {
			return &d.Rules[i]
		}
	}
	return nil
}

// GetRuleByNumber finds a rule by its number.
func (d *Document) GetRuleByNumber(num int) *Rule {
	for i := range d.Rules {
		if d.Rules[i].Number == num {
			return &d.Rules[i]
		}
	}
	return nil
}
