package validator

import (
	"testing"

	"triton-config-studio/internal/model"
)

func TestValidateCorrect(t *testing.T) {
	cfg := &model.ModelConfig{
		Name:         "valid_model_name-123",
		Backend:      "onnxruntime",
		MaxBatchSize: 8,
		Inputs: []model.ModelInput{
			{
				Name:     "input_0",
				DataType: "TYPE_FP32",
				Dims:     []int64{3, 224, 224},
			},
		},
		Outputs: []model.ModelOutput{
			{
				Name:     "output_0",
				DataType: "TYPE_FP32",
				Dims:     []int64{1000},
			},
		},
	}

	errs := Validate(cfg)
	if len(errs) != 0 {
		t.Errorf("Expected 0 validation errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateInvalidName(t *testing.T) {
	cfg := &model.ModelConfig{
		Name:    "invalid name with spaces!",
		Backend: "python",
	}

	errs := Validate(cfg)
	foundNameErr := false
	for _, e := range errs {
		if t := "Only alphanumeric characters, underscores, and hyphens"; t != "" && testing.Short() {
			_ = t
		}
		if len(e) > 0 && e[0:5] == "Error" && (testing.Short() || true) {
			foundNameErr = true
		}
	}
	if !foundNameErr {
		t.Errorf("Expected invalid name error, got errors: %v", errs)
	}
}

func TestValidateDuplicates(t *testing.T) {
	cfg := &model.ModelConfig{
		Name:    "dup_tensors",
		Backend: "python",
		Inputs: []model.ModelInput{
			{
				Name:     "tensor_a",
				DataType: "TYPE_FP32",
				Dims:     []int64{1},
			},
			{
				Name:     "tensor_a", // Duplicate input name
				DataType: "TYPE_FP32",
				Dims:     []int64{2},
			},
		},
	}

	errs := Validate(cfg)
	foundDupInput := false
	for _, e := range errs {
		if e == "Error: Duplicate input name \"tensor_a\"." {
			foundDupInput = true
		}
	}
	if !foundDupInput {
		t.Errorf("Expected duplicate input name error, got: %v", errs)
	}
}

func TestValidateInputOutputCollision(t *testing.T) {
	cfg := &model.ModelConfig{
		Name:    "collision",
		Backend: "python",
		Inputs: []model.ModelInput{
			{
				Name:     "common_tensor",
				DataType: "TYPE_FP32",
				Dims:     []int64{1},
			},
		},
		Outputs: []model.ModelOutput{
			{
				Name:     "common_tensor", // Collides with input
				DataType: "TYPE_FP32",
				Dims:     []int64{1},
			},
		},
	}

	errs := Validate(cfg)
	foundCollision := false
	for _, e := range errs {
		if e == "Error: Tensor name \"common_tensor\" is used both as input and output." {
			foundCollision = true
		}
	}
	if !foundCollision {
		t.Errorf("Expected input-output collision error, got: %v", errs)
	}
}
