// Package checker provides the checker registry and execution framework.
package checker

import (
	"fmt"
	"sort"
	"sync"

	"github.com/michaellady/agents-lint/internal/transcript"
)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Checker)
)

// Register adds a checker to the global registry.
// It panics if a checker with the same ID is already registered.
func Register(c Checker) {
	registryMu.Lock()
	defer registryMu.Unlock()

	id := c.ID()
	if _, exists := registry[id]; exists {
		panic(fmt.Sprintf("checker already registered: %s", id))
	}
	registry[id] = c
}

// GetByID returns a checker by its ID, or nil if not found.
func GetByID(id string) Checker {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[id]
}

// GetAll returns all registered checkers sorted by ID.
func GetAll() []Checker {
	registryMu.RLock()
	defer registryMu.RUnlock()

	checkers := make([]Checker, 0, len(registry))
	for _, c := range registry {
		checkers = append(checkers, c)
	}

	sort.Slice(checkers, func(i, j int) bool {
		return checkers[i].ID() < checkers[j].ID()
	})

	return checkers
}

// IDs returns the IDs of all registered checkers sorted alphabetically.
func IDs() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// RunAll executes all registered checkers against a transcript.
func RunAll(t *transcript.Transcript) *Result {
	checkers := GetAll()
	return Run(t, checkers)
}

// Run executes the specified checkers against a transcript.
func Run(t *transcript.Transcript, checkers []Checker) *Result {
	result := &Result{
		CheckersRun: make([]string, 0, len(checkers)),
		Violations:  make([]Violation, 0),
	}

	for _, c := range checkers {
		result.CheckersRun = append(result.CheckersRun, c.ID())
		violations := c.Check(t)
		result.Violations = append(result.Violations, violations...)
	}

	return result
}

// RunByIDs executes checkers with the given IDs against a transcript.
// Unknown IDs are silently ignored.
func RunByIDs(t *transcript.Transcript, ids []string) *Result {
	var checkers []Checker
	for _, id := range ids {
		if c := GetByID(id); c != nil {
			checkers = append(checkers, c)
		}
	}
	return Run(t, checkers)
}

// Clear removes all registered checkers. Intended for testing only.
func Clear() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]Checker)
}
