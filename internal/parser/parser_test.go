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

func TestParseTritonSchemaFieldNames(t *testing.T) {
	content := `
	name: "schema_model"
	backend: "onnxruntime"
	max_batch_size: 8
	input {
		name: "input_ids"
		data_type: TYPE_INT32
		dims: [ 1 ]
		reshape { shape: [ ] }
	}
	output {
		name: "logits"
		data_type: TYPE_FP32
		dims: [ -1, 768 ]
		reshape { shape: [ 768 ] }
	}
	dynamic_batching {
		default_queue_policy {
			timeout_action: DELAY
			default_timeout_microseconds: 5000
			max_queue_size: 16
		}
		priority_queue_policy {
			key: 2
			value {
				timeout_action: REJECT
				default_timeout_microseconds: 1000
				max_queue_size: 8
			}
		}
	}
	optimization {
		input_pinned_memory { enable: true }
		output_pinned_memory { enable: true }
	}
	model_warmup {
		name: "sample"
		batch_size: 1
		inputs {
			key: "input_ids"
			value {
				data_type: TYPE_INT32
				dims: [ 1 ]
				zero_data: true
			}
		}
	}
	`

	cfg, err := Parse(content)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Inputs[0].Reshape == nil || len(cfg.Inputs[0].Reshape.Dims) != 0 {
		t.Fatalf("expected scalar input reshape shape [], got %#v", cfg.Inputs[0].Reshape)
	}
	if cfg.Outputs[0].Reshape == nil || len(cfg.Outputs[0].Reshape.Dims) != 1 || cfg.Outputs[0].Reshape.Dims[0] != 768 {
		t.Fatalf("expected output reshape shape [768], got %#v", cfg.Outputs[0].Reshape)
	}
	if cfg.DynamicBatching == nil || cfg.DynamicBatching.DefaultQueuePolicy == nil {
		t.Fatal("expected default queue policy")
	}
	if cfg.DynamicBatching.DefaultQueuePolicy.Action != "DELAY" || cfg.DynamicBatching.DefaultQueuePolicy.TimeoutMicroseconds != 5000 || cfg.DynamicBatching.DefaultQueuePolicy.MaxQueueSize != 16 {
		t.Fatalf("default queue policy mismatch: %#v", cfg.DynamicBatching.DefaultQueuePolicy)
	}
	if len(cfg.DynamicBatching.PriorityQueuePolicy) != 1 {
		t.Fatalf("expected one priority queue policy, got %#v", cfg.DynamicBatching.PriorityQueuePolicy)
	}
	priorityPolicy := cfg.DynamicBatching.PriorityQueuePolicy[0]
	if priorityPolicy.Priority != 2 || priorityPolicy.QueuePolicy == nil || priorityPolicy.QueuePolicy.Action != "REJECT" || priorityPolicy.QueuePolicy.TimeoutMicroseconds != 1000 || priorityPolicy.QueuePolicy.MaxQueueSize != 8 {
		t.Fatalf("priority queue policy mismatch: %#v", priorityPolicy)
	}
	if cfg.Optimization == nil || !cfg.Optimization.InputPinnedMemory || !cfg.Optimization.OutputPinnedMemory {
		t.Fatalf("expected pinned memory settings to parse, got %#v", cfg.Optimization)
	}
	if len(cfg.Warmups) != 1 || len(cfg.Warmups[0].Inputs) != 1 || cfg.Warmups[0].Inputs[0].Key != "input_ids" {
		t.Fatalf("warmup inputs did not parse: %#v", cfg.Warmups)
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
