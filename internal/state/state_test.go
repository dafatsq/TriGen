package state

import (
	"testing"

	"triton-config-studio/internal/model"
)

func TestSetConfigClearsUIErrors(t *testing.T) {
	s := NewAppState()
	s.SetUIError("field", "Error: stale invalid input")

	s.SetConfig(&model.ModelConfig{Name: "fresh"})

	if errs := s.GetUIErrors(); len(errs) != 0 {
		t.Fatalf("expected SetConfig to clear stale UI errors, got %v", errs)
	}
}
