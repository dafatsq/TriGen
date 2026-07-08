package generator

import (
	"strings"
	"testing"

	"triton-config-studio/internal/model"
)

func TestGenerateUsesTritonSchemaFieldNames(t *testing.T) {
	cfg := &model.ModelConfig{
		Name:         "schema_model",
		Backend:      "onnxruntime",
		MaxBatchSize: 8,
		Inputs: []model.ModelInput{
			{
				Name:     "input_ids",
				DataType: "TYPE_INT32",
				Dims:     []int64{1},
				Reshape:  &model.Reshape{Dims: []int64{}},
			},
		},
		Outputs: []model.ModelOutput{
			{
				Name:     "logits",
				DataType: "TYPE_FP32",
				Dims:     []int64{-1, 768},
				Reshape:  &model.Reshape{Dims: []int64{768}},
			},
		},
		DynamicBatching: &model.DynamicBatching{
			DefaultQueuePolicy: &model.QueuePolicy{
				TimeoutMicroseconds: 5000,
				MaxQueueSize:        16,
				Action:              "REJECT",
			},
			PriorityQueuePolicy: []model.PriorityQueuePolicy{
				{
					Priority: 1,
					QueuePolicy: &model.QueuePolicy{
						TimeoutMicroseconds: 1000,
						Action:              "DELAY",
					},
				},
			},
		},
		Optimization: &model.Optimization{
			InputPinnedMemory:  true,
			OutputPinnedMemory: true,
		},
	}

	got := Generate(cfg)

	for _, want := range []string{
		"shape: [ 768 ]",
		"default_timeout_microseconds: 5000",
		"timeout_action: REJECT",
		"default_timeout_microseconds: 1000",
		"timeout_action: DELAY",
		"input_pinned_memory {\n    enable: true\n  }",
		"output_pinned_memory {\n    enable: true\n  }",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated config missing %q:\n%s", want, got)
		}
	}

	legacyLines := map[string]bool{
		"timeout_microseconds: 5000": true,
		"timeout_microseconds: 1000": true,
		"action: REJECT":             true,
		"action: DELAY":              true,
		"input_pinned_memory: true":  true,
		"output_pinned_memory: true": true,
	}
	for _, line := range strings.Split(got, "\n") {
		if legacyLines[strings.TrimSpace(line)] {
			t.Fatalf("generated config contains legacy line %q:\n%s", strings.TrimSpace(line), got)
		}
	}
}
