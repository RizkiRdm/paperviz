package services

import (
	"strings"
	"testing"
)

func TestSimplifiedPromptContainsAllSections(t *testing.T) {
	sections := []string{
		"Research Question",
		"Method",
		"Main Findings",
		"Evidence",
		"Limitations",
		"Key Figures",
		"Key Tables",
		"Conclusion",
	}
	for _, section := range sections {
		if !strings.Contains(simplifiedPrompt, "## "+section) {
			t.Errorf("simplifiedPrompt missing section header: ## %s", section)
		}
	}
}

func TestSimplifiedPromptPreservesFactualIntegrityInstruction(t *testing.T) {
	if !strings.Contains(simplifiedPrompt, "Do NOT remove, change, round, or invent any number") {
		t.Error("simplifiedPrompt missing factual integrity instruction")
	}
}
