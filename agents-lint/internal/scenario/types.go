// Package scenario defines test scenario types and loading.
package scenario

// Scenario defines a test case for agents-test.
type Scenario struct {
	// Name is a short identifier for the scenario.
	Name string `yaml:"name"`

	// Description explains what this scenario tests.
	Description string `yaml:"description"`

	// Prompt is the user prompt to send to Claude.
	Prompt string `yaml:"prompt"`

	// SystemPrompt is additional system prompt content (optional).
	// AGENTS.md content is always appended.
	SystemPrompt string `yaml:"system_prompt,omitempty"`

	// WorkDir is the working directory for the test (optional).
	// If not specified, a temp directory is created.
	WorkDir string `yaml:"work_dir,omitempty"`

	// SetupCommands are shell commands to run before the test.
	SetupCommands []string `yaml:"setup,omitempty"`

	// CleanupCommands are shell commands to run after the test.
	CleanupCommands []string `yaml:"cleanup,omitempty"`

	// ExpectedCheckers lists checker IDs that should find violations.
	// Used to validate test scenarios work correctly.
	ExpectViolations []string `yaml:"expect_violations,omitempty"`

	// ExpectPass means the scenario should pass all checkers.
	ExpectPass bool `yaml:"expect_pass,omitempty"`

	// Model overrides the default model (optional).
	Model string `yaml:"model,omitempty"`

	// MaxTurns limits the number of turns (optional).
	MaxTurns int `yaml:"max_turns,omitempty"`

	// Timeout in seconds (default 120).
	Timeout int `yaml:"timeout,omitempty"`

	// Tags for filtering scenarios.
	Tags []string `yaml:"tags,omitempty"`
}

// Result contains the outcome of running a scenario.
type Result struct {
	// Scenario is the scenario that was run.
	Scenario *Scenario

	// TranscriptPath is where the NDJSON output was saved.
	TranscriptPath string

	// Passed is true if the scenario met expectations.
	Passed bool

	// Violations found during checking.
	ViolationCount int

	// Error if the scenario failed to run.
	Error error

	// DurationMS is how long the scenario took.
	DurationMS int64
}
