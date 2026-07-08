package model

import "testing"

func TestCloneConfigDeepCopiesNestedState(t *testing.T) {
	original := &ModelConfig{
		Name: "template",
		Inputs: []ModelInput{
			{Name: "input", Dims: []int64{1, 2, 3}, Reshape: &Reshape{Dims: []int64{6}}},
		},
		DynamicBatching: &DynamicBatching{
			DefaultQueuePolicy: &QueuePolicy{TimeoutMicroseconds: 100, Action: "REJECT"},
		},
	}

	clone := CloneConfig(original)
	if clone == nil {
		t.Fatal("CloneConfig returned nil for non-nil config")
	}

	clone.Name = "edited"
	clone.Inputs[0].Name = "edited_input"
	clone.Inputs[0].Dims[0] = 99
	clone.Inputs[0].Reshape.Dims[0] = 42
	clone.DynamicBatching.DefaultQueuePolicy.Action = "DELAY"

	if original.Name != "template" {
		t.Fatalf("original name changed to %q", original.Name)
	}
	if original.Inputs[0].Name != "input" {
		t.Fatalf("original input name changed to %q", original.Inputs[0].Name)
	}
	if original.Inputs[0].Dims[0] != 1 {
		t.Fatalf("original input dims changed to %v", original.Inputs[0].Dims)
	}
	if original.Inputs[0].Reshape.Dims[0] != 6 {
		t.Fatalf("original reshape dims changed to %v", original.Inputs[0].Reshape.Dims)
	}
	if original.DynamicBatching.DefaultQueuePolicy.Action != "REJECT" {
		t.Fatalf("original queue policy action changed to %q", original.DynamicBatching.DefaultQueuePolicy.Action)
	}
}

func TestCloneConfigNil(t *testing.T) {
	if CloneConfig(nil) != nil {
		t.Fatal("CloneConfig(nil) should return nil")
	}
}
