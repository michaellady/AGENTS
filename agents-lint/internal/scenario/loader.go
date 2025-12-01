package scenario

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads a scenario from a YAML file.
func Load(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario file: %w", err)
	}

	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario YAML: %w", err)
	}

	// Set defaults
	if s.Timeout == 0 {
		s.Timeout = 120
	}

	return &s, nil
}

// LoadAll reads all scenarios from a directory.
func LoadAll(dir string) ([]*Scenario, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read scenario directory: %w", err)
	}

	var scenarios []*Scenario
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 5 || name[len(name)-5:] != ".yaml" {
			continue
		}

		s, err := Load(dir + "/" + name)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", name, err)
		}
		scenarios = append(scenarios, s)
	}

	return scenarios, nil
}
