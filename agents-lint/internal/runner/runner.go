// Package runner executes Claude Code and captures output.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/michaellady/agents-lint/internal/checker"
	"github.com/michaellady/agents-lint/internal/scenario"
	"github.com/michaellady/agents-lint/internal/transcript"
)

// Config holds runner configuration.
type Config struct {
	// ClaudePath is the path to the claude binary.
	ClaudePath string

	// AGENTSPath is the path to AGENTS.md file.
	AGENTSPath string

	// OutputDir is where to save transcripts.
	OutputDir string

	// Model overrides the scenario's model.
	Model string

	// DryRun prints the command without executing.
	DryRun bool

	// Verbose enables detailed output.
	Verbose bool
}

// Runner executes test scenarios.
type Runner struct {
	config Config
}

// New creates a new Runner with the given config.
func New(cfg Config) *Runner {
	if cfg.ClaudePath == "" {
		cfg.ClaudePath = "claude"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "."
	}
	return &Runner{config: cfg}
}

// Run executes a scenario and returns the result.
func (r *Runner) Run(ctx context.Context, s *scenario.Scenario) *scenario.Result {
	result := &scenario.Result{Scenario: s}
	start := time.Now()

	defer func() {
		result.DurationMS = time.Since(start).Milliseconds()
	}()

	// Build command
	args := []string{
		"-p",
		"--verbose",
		"--output-format", "stream-json",
	}

	// Add model
	model := s.Model
	if r.config.Model != "" {
		model = r.config.Model
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	// Add max turns
	if s.MaxTurns > 0 {
		args = append(args, "--max-turns", fmt.Sprintf("%d", s.MaxTurns))
	}

	// Add system prompt with AGENTS.md
	systemPrompt := s.SystemPrompt
	if r.config.AGENTSPath != "" {
		agentsContent, err := os.ReadFile(r.config.AGENTSPath)
		if err != nil {
			result.Error = fmt.Errorf("read AGENTS.md: %w", err)
			return result
		}
		if systemPrompt != "" {
			systemPrompt += "\n\n"
		}
		systemPrompt += string(agentsContent)
	}
	if systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}

	// Add the prompt
	args = append(args, s.Prompt)

	if r.config.DryRun {
		fmt.Printf("Would run: %s %v\n", r.config.ClaudePath, args)
		return result
	}

	// Determine working directory
	workDir := s.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "agents-test-*")
		if err != nil {
			result.Error = fmt.Errorf("create temp dir: %w", err)
			return result
		}
		defer os.RemoveAll(workDir)
	}

	// Run setup commands
	for _, cmd := range s.SetupCommands {
		if err := runShell(ctx, workDir, cmd); err != nil {
			result.Error = fmt.Errorf("setup command %q: %w", cmd, err)
			return result
		}
	}

	// Execute claude
	timeout := time.Duration(s.Timeout) * time.Second
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, r.config.ClaudePath, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if r.config.Verbose {
		fmt.Printf("Running: %s %v\n", r.config.ClaudePath, args)
		fmt.Printf("WorkDir: %s\n", workDir)
	}

	err := cmd.Run()
	if err != nil && cmdCtx.Err() == context.DeadlineExceeded {
		result.Error = fmt.Errorf("timeout after %ds", s.Timeout)
		return result
	}
	// Note: claude may return non-zero exit codes for various reasons,
	// but we still want to check the transcript

	// Run cleanup commands
	for _, cmd := range s.CleanupCommands {
		_ = runShell(ctx, workDir, cmd) // Best effort cleanup
	}

	// Save transcript
	transcriptName := fmt.Sprintf("%s-%d.ndjson", s.Name, time.Now().Unix())
	result.TranscriptPath = filepath.Join(r.config.OutputDir, transcriptName)
	if err := os.WriteFile(result.TranscriptPath, stdout.Bytes(), 0644); err != nil {
		result.Error = fmt.Errorf("save transcript: %w", err)
		return result
	}

	// Parse and check transcript
	t, err := transcript.ParseBytes(stdout.Bytes())
	if err != nil {
		result.Error = fmt.Errorf("parse transcript: %w", err)
		return result
	}

	checkResult := checker.RunAll(t)
	result.ViolationCount = len(checkResult.Violations)

	// Determine if passed based on expectations
	if s.ExpectPass {
		result.Passed = !checkResult.HasErrors()
	} else if len(s.ExpectViolations) > 0 {
		// Check that expected violations were found
		foundViolations := make(map[string]bool)
		for _, v := range checkResult.Violations {
			foundViolations[v.CheckerID] = true
		}
		result.Passed = true
		for _, expected := range s.ExpectViolations {
			if !foundViolations[expected] {
				result.Passed = false
				break
			}
		}
	} else {
		// Default: pass if no errors
		result.Passed = !checkResult.HasErrors()
	}

	return result
}

// runShell executes a shell command.
func runShell(ctx context.Context, dir, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	return cmd.Run()
}
