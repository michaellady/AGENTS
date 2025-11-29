package scenario

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadAll(t *testing.T) {
	// Get the scenarios directory relative to this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	scenariosDir := filepath.Join(filepath.Dir(filename), "..", "..", "scenarios")

	scenarios, err := LoadAll(scenariosDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	if len(scenarios) == 0 {
		t.Fatal("expected at least one scenario")
	}

	// Verify each scenario has required fields
	for _, s := range scenarios {
		if s.Name == "" {
			t.Error("scenario missing name")
		}
		if s.Prompt == "" {
			t.Errorf("scenario %s missing prompt", s.Name)
		}
		if s.Timeout == 0 {
			t.Errorf("scenario %s has zero timeout (should default to 120)", s.Name)
		}
	}

	t.Logf("Loaded %d scenarios", len(scenarios))
}

func TestLoad(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	scenariosDir := filepath.Join(filepath.Dir(filename), "..", "..", "scenarios")

	s, err := Load(filepath.Join(scenariosDir, "simple-read.yaml"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if s.Name != "simple-read" {
		t.Errorf("expected name 'simple-read', got %s", s.Name)
	}
	if !s.ExpectPass {
		t.Error("expected ExpectPass to be true")
	}
	if s.Timeout != 60 {
		t.Errorf("expected timeout 60, got %d", s.Timeout)
	}
}

func TestScenarioDefaults(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get test file path")
	}
	scenariosDir := filepath.Join(filepath.Dir(filename), "..", "..", "scenarios")

	scenarios, err := LoadAll(scenariosDir)
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	for _, s := range scenarios {
		// All scenarios should have timeout set (either explicit or default)
		if s.Timeout == 0 {
			t.Errorf("scenario %s has zero timeout", s.Name)
		}
	}
}
