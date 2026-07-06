package parser

import (
	"testing"
)

func TestParseSimple(t *testing.T) {
	content := `
	# This is a comment
	name: "my_model"
	platform: "tensorrt_plan"
	max_batch_size: 8

	input [
		{
			name: "input_0"
			data_type: TYPE_FP32
			dims: [ 3, 224, 224 ]
		}
	]

	output {
		name: "output_0"
		data_type: TYPE_FP32
		dims: [ 1000 ]
	}

	parameters {
		key: "tokenizer_dir"
		value: {
			string_value: "./tokenizer"
		}
	}
	`

	cfg, err := Parse(content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if cfg.Name != "my_model" {
		t.Errorf("Expected name to be my_model, got %q", cfg.Name)
	}
	if cfg.Platform != "tensorrt_plan" {
		t.Errorf("Expected platform to be tensorrt_plan, got %q", cfg.Platform)
	}
	if cfg.MaxBatchSize != 8 {
		t.Errorf("Expected max_batch_size to be 8, got %d", cfg.MaxBatchSize)
	}

	if len(cfg.Inputs) != 1 {
		t.Fatalf("Expected 1 input, got %d", len(cfg.Inputs))
	}
	in := cfg.Inputs[0]
	if in.Name != "input_0" {
		t.Errorf("Expected input name to be input_0, got %q", in.Name)
	}
	if in.DataType != "TYPE_FP32" {
		t.Errorf("Expected input data_type to be TYPE_FP32, got %q", in.DataType)
	}
	if len(in.Dims) != 3 || in.Dims[0] != 3 || in.Dims[1] != 224 || in.Dims[2] != 224 {
		t.Errorf("Expected input dims to be [3, 224, 224], got %v", in.Dims)
	}

	if len(cfg.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(cfg.Outputs))
	}
	out := cfg.Outputs[0]
	if out.Name != "output_0" {
		t.Errorf("Expected output name to be output_0, got %q", out.Name)
	}

	if len(cfg.Parameters) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(cfg.Parameters))
	}
	param := cfg.Parameters[0]
	if param.Key != "tokenizer_dir" {
		t.Errorf("Expected parameter key tokenizer_dir, got %q", param.Key)
	}
	if param.Value.StringValue != "./tokenizer" {
		t.Errorf("Expected parameter value string_value ./tokenizer, got %q", param.Value.StringValue)
	}
}

func TestParseComplex(t *testing.T) {
	content := `
	name: "preprocessing"
	backend: "python"
	max_batch_size: 128
	input [
		{
			name: "QUERY"
			data_type: TYPE_STRING
			dims: [ 1 ]
		},
		{
			name: "IMAGE"
			data_type: TYPE_FP16
			dims: [ 3, 224, 224 ]
			optional: true
		}
	]
	output [
		{
			name: "INPUT_ID"
			data_type: TYPE_INT32
			dims: [ -1 ]
		}
	]
	instance_group [
		{
			count: 4
			kind: KIND_CPU
		}
	]
	dynamic_batching {
		preferred_batch_size: [ 4, 8, 16 ]
		max_queue_delay_microseconds: 5000
	}
	`

	cfg, err := Parse(content)
	if err != nil {
		t.Fatalf("Failed to parse complex: %v", err)
	}

	if len(cfg.Inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(cfg.Inputs))
	}
	if cfg.Inputs[1].Optional != true {
		t.Errorf("Expected input[1] to be optional")
	}
	if len(cfg.InstanceGroups) != 1 || cfg.InstanceGroups[0].Count != 4 || cfg.InstanceGroups[0].Kind != "KIND_CPU" {
		t.Errorf("Instance group mismatch")
	}
	if cfg.DynamicBatching == nil || len(cfg.DynamicBatching.PreferredBatchSize) != 3 || cfg.DynamicBatching.MaxQueueDelayMicroseconds != 5000 {
		t.Errorf("Dynamic batching mismatch")
	}
}
