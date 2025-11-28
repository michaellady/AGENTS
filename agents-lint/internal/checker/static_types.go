package checker

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/michaellady/agents-lint/internal/transcript"
)

func init() {
	Register(&StaticTypes{})
}

// StaticTypes checks that new code uses statically typed languages.
// Per AGENTS.md Rule 9: "Prefer Static Types"
// - New projects: Use Go, Kotlin, TypeScript, or Rust
// - Scripting: Always use type hints (Python) or TypeScript (not JS)
type StaticTypes struct{}

func (c *StaticTypes) ID() string {
	return "static-types"
}

func (c *StaticTypes) Description() string {
	return "Ensures new code uses TypeScript instead of JavaScript (Rule 9)"
}

// configFilePatterns are JS files that are exceptions (config files typically require .js)
var configFilePatterns = []string{
	".config.js",
	".config.mjs",
	".config.cjs",
	"rc.js",
	"rc.mjs",
	"rc.cjs",
}

// isConfigFile checks if a file path is a known config file that requires .js
func isConfigFile(path string) bool {
	base := filepath.Base(path)

	// Check common config file patterns
	for _, pattern := range configFilePatterns {
		if strings.HasSuffix(base, pattern) {
			return true
		}
	}

	// Check for dotfile configs like .eslintrc.js
	if strings.HasPrefix(base, ".") && strings.HasSuffix(base, ".js") {
		return true
	}

	return false
}

func (c *StaticTypes) Check(t *transcript.Transcript) []Violation {
	var violations []Violation

	for _, tc := range t.ToolCalls {
		// Only check Write tool (creating new files)
		// Edit tool is for modifying existing files, which is OK for legacy code
		if tc.Name != "Write" {
			continue
		}

		// Parse the input to get file_path
		var input struct {
			FilePath string `json:"file_path"`
		}
		if err := json.Unmarshal(tc.Input, &input); err != nil {
			continue
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(input.FilePath))

		// Flag .js and .jsx files (should use .ts and .tsx)
		if ext == ".js" || ext == ".jsx" {
			// Skip config files which often require .js
			if isConfigFile(input.FilePath) {
				continue
			}

			suggestion := ".ts"
			if ext == ".jsx" {
				suggestion = ".tsx"
			}

			violations = append(violations, Violation{
				CheckerID:  c.ID(),
				Rule:       "Rule 9",
				Severity:   SeverityWarning,
				Message:    "Creating " + ext + " file; prefer " + suggestion + " for type safety",
				EventUUID:  tc.EventUUID,
				ToolCallID: tc.ID,
			})
		}
	}

	return violations
}
