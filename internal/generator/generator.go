package generator

import (
	"fmt"
	"strings"

	"triton-config-studio/internal/model"
)

func formatDims(dims []int64) string {
	var parts []string
	for _, d := range dims {
		parts = append(parts, fmt.Sprintf("%d", d))
	}
	return strings.Join(parts, ", ")
}

func formatInt64List(dims []int64) string {
	if len(dims) == 0 {
		return "[ ]"
	}
	return fmt.Sprintf("[ %s ]", formatDims(dims))
}

func formatInt32Slice(slice []int32) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, ", ")
}

func formatInt64Slice(slice []int64) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, ", ")
}

func writeIndent(sb *strings.Builder, level int, format string, args ...interface{}) {
	indent := strings.Repeat("  ", level)
	sb.WriteString(indent + fmt.Sprintf(format, args...) + "\n")
}

// Generate formats a ModelConfig struct into a config.pbtxt string
func Generate(cfg *model.ModelConfig) string {
	var sb strings.Builder

	if cfg.Name != "" {
		writeIndent(&sb, 0, "name: %q", cfg.Name)
	}
	if cfg.Platform != "" {
		writeIndent(&sb, 0, "platform: %q", cfg.Platform)
	}
	if cfg.Backend != "" {
		writeIndent(&sb, 0, "backend: %q", cfg.Backend)
	}
	if cfg.MaxBatchSize >= 0 {
		writeIndent(&sb, 0, "max_batch_size: %d", cfg.MaxBatchSize)
	}
	if cfg.DefaultModelFilename != "" {
		writeIndent(&sb, 0, "default_model_filename: %q", cfg.DefaultModelFilename)
	}

	// Inputs
	for _, in := range cfg.Inputs {
		writeIndent(&sb, 0, "input {")
		writeIndent(&sb, 1, "name: %q", in.Name)
		writeIndent(&sb, 1, "data_type: %s", in.DataType)
		writeIndent(&sb, 1, "dims: [ %s ]", formatDims(in.Dims))
		if in.Optional {
			writeIndent(&sb, 1, "optional: true")
		}
		if in.AllowRaggedBatch {
			writeIndent(&sb, 1, "allow_ragged_batch: true")
		}
		if in.IsShapeTensor {
			writeIndent(&sb, 1, "is_shape_tensor: true")
		}
		if in.Reshape != nil {
			writeIndent(&sb, 1, "reshape {")
			writeIndent(&sb, 2, "shape: %s", formatInt64List(in.Reshape.Dims))
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Outputs
	for _, out := range cfg.Outputs {
		writeIndent(&sb, 0, "output {")
		writeIndent(&sb, 1, "name: %q", out.Name)
		writeIndent(&sb, 1, "data_type: %s", out.DataType)
		writeIndent(&sb, 1, "dims: [ %s ]", formatDims(out.Dims))
		if out.LabelFilename != "" {
			writeIndent(&sb, 1, "label_filename: %q", out.LabelFilename)
		}
		if out.Reshape != nil {
			writeIndent(&sb, 1, "reshape {")
			writeIndent(&sb, 2, "shape: %s", formatInt64List(out.Reshape.Dims))
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Version Policy
	if cfg.VersionPolicy != nil {
		writeIndent(&sb, 0, "version_policy {")
		if cfg.VersionPolicy.Latest != nil {
			writeIndent(&sb, 1, "latest {")
			writeIndent(&sb, 2, "num_versions: %d", cfg.VersionPolicy.Latest.NumVersions)
			writeIndent(&sb, 1, "}")
		} else if cfg.VersionPolicy.All != nil {
			writeIndent(&sb, 1, "all {}")
		} else if cfg.VersionPolicy.Specific != nil {
			writeIndent(&sb, 1, "specific {")
			writeIndent(&sb, 2, "versions: [ %s ]", formatInt64Slice(cfg.VersionPolicy.Specific.Versions))
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Instance Groups
	for _, grp := range cfg.InstanceGroups {
		writeIndent(&sb, 0, "instance_group {")
		if grp.Count > 0 {
			writeIndent(&sb, 1, "count: %d", grp.Count)
		}
		if grp.Kind != "" {
			writeIndent(&sb, 1, "kind: %s", grp.Kind)
		}
		if len(grp.Gpus) > 0 {
			writeIndent(&sb, 1, "gpus: [ %s ]", formatInt32Slice(grp.Gpus))
		}
		if grp.HostPolicy != "" {
			writeIndent(&sb, 1, "host_policy: %q", grp.HostPolicy)
		}
		writeIndent(&sb, 0, "}")
	}

	// Dynamic Batching
	if cfg.DynamicBatching != nil {
		writeIndent(&sb, 0, "dynamic_batching {")
		if len(cfg.DynamicBatching.PreferredBatchSize) > 0 {
			writeIndent(&sb, 1, "preferred_batch_size: [ %s ]", formatInt32Slice(cfg.DynamicBatching.PreferredBatchSize))
		}
		if cfg.DynamicBatching.MaxQueueDelayMicroseconds > 0 {
			writeIndent(&sb, 1, "max_queue_delay_microseconds: %d", cfg.DynamicBatching.MaxQueueDelayMicroseconds)
		}
		if cfg.DynamicBatching.PreserveOrdering {
			writeIndent(&sb, 1, "preserve_ordering: true")
		}
		if cfg.DynamicBatching.PriorityLevels > 0 {
			writeIndent(&sb, 1, "priority_levels: %d", cfg.DynamicBatching.PriorityLevels)
		}
		if cfg.DynamicBatching.DefaultQueuePolicy != nil {
			writeIndent(&sb, 1, "default_queue_policy {")
			qp := cfg.DynamicBatching.DefaultQueuePolicy
			if qp.TimeoutMicroseconds > 0 {
				writeIndent(&sb, 2, "default_timeout_microseconds: %d", qp.TimeoutMicroseconds)
			}
			if qp.MaxQueueSize > 0 {
				writeIndent(&sb, 2, "max_queue_size: %d", qp.MaxQueueSize)
			}
			if qp.Action != "" {
				writeIndent(&sb, 2, "timeout_action: %s", qp.Action)
			}
			writeIndent(&sb, 1, "}")
		}
		for _, pqp := range cfg.DynamicBatching.PriorityQueuePolicy {
			writeIndent(&sb, 1, "priority_queue_policy {")
			writeIndent(&sb, 2, "key: %d", pqp.Priority)
			writeIndent(&sb, 2, "value {")
			if pqp.QueuePolicy != nil {
				qp := pqp.QueuePolicy
				if qp.TimeoutMicroseconds > 0 {
					writeIndent(&sb, 3, "default_timeout_microseconds: %d", qp.TimeoutMicroseconds)
				}
				if qp.MaxQueueSize > 0 {
					writeIndent(&sb, 3, "max_queue_size: %d", qp.MaxQueueSize)
				}
				if qp.Action != "" {
					writeIndent(&sb, 3, "timeout_action: %s", qp.Action)
				}
			}
			writeIndent(&sb, 2, "}")
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Sequence Batching
	if cfg.SequenceBatching != nil {
		writeIndent(&sb, 0, "sequence_batching {")
		if cfg.SequenceBatching.MaxSequenceIdleMicroseconds > 0 {
			writeIndent(&sb, 1, "max_sequence_idle_microseconds: %d", cfg.SequenceBatching.MaxSequenceIdleMicroseconds)
		}
		if cfg.SequenceBatching.Direct != nil {
			writeIndent(&sb, 1, "direct {")
			if cfg.SequenceBatching.Direct.MinimumSlotAllocation > 0 {
				writeIndent(&sb, 2, "minimum_slot_allocation: %d", cfg.SequenceBatching.Direct.MinimumSlotAllocation)
			}
			writeIndent(&sb, 1, "}")
		}
		if cfg.SequenceBatching.Oldest != nil {
			writeIndent(&sb, 1, "oldest {")
			if cfg.SequenceBatching.Oldest.MaxQueueDelayMicroseconds > 0 {
				writeIndent(&sb, 2, "max_queue_delay_microseconds: %d", cfg.SequenceBatching.Oldest.MaxQueueDelayMicroseconds)
			}
			writeIndent(&sb, 1, "}")
		}
		for _, ci := range cfg.SequenceBatching.ControlInputs {
			writeIndent(&sb, 1, "control_input {")
			writeIndent(&sb, 2, "name: %q", ci.Name)
			for _, ctrl := range ci.Controls {
				writeIndent(&sb, 2, "control {")
				writeIndent(&sb, 3, "kind: %s", ctrl.Kind)
				if len(ctrl.Int32Value) > 0 {
					writeIndent(&sb, 3, "int32_value: [ %s ]", formatInt32Slice(ctrl.Int32Value))
				}
				if len(ctrl.Fp32Value) > 0 {
					var fpParts []string
					for _, f := range ctrl.Fp32Value {
						fpParts = append(fpParts, fmt.Sprintf("%g", f))
					}
					writeIndent(&sb, 3, "fp32_value: [ %s ]", strings.Join(fpParts, ", "))
				}
				writeIndent(&sb, 2, "}")
			}
			writeIndent(&sb, 1, "}")
		}
		for _, s := range cfg.SequenceBatching.States {
			writeIndent(&sb, 1, "state {")
			writeIndent(&sb, 2, "name: %q", s.Name)
			writeIndent(&sb, 2, "data_type: %s", s.DataType)
			writeIndent(&sb, 2, "dims: [ %s ]", formatDims(s.Dims))
			for _, init := range s.InitialState {
				writeIndent(&sb, 2, "initial_state {")
				if init.Name != "" {
					writeIndent(&sb, 3, "name: %q", init.Name)
				}
				writeIndent(&sb, 3, "data_type: %s", init.DataType)
				writeIndent(&sb, 3, "dims: [ %s ]", formatDims(init.Dims))
				if init.ZeroData {
					writeIndent(&sb, 3, "zero_data: true")
				}
				writeIndent(&sb, 2, "}")
			}
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Optimization
	if cfg.Optimization != nil {
		hasContent := cfg.Optimization.InputPinnedMemory || cfg.Optimization.OutputPinnedMemory || cfg.Optimization.ExecutionAccelerators != nil
		if hasContent {
			writeIndent(&sb, 0, "optimization {")
			if cfg.Optimization.InputPinnedMemory {
				writeIndent(&sb, 1, "input_pinned_memory {")
				writeIndent(&sb, 2, "enable: true")
				writeIndent(&sb, 1, "}")
			}
			if cfg.Optimization.OutputPinnedMemory {
				writeIndent(&sb, 1, "output_pinned_memory {")
				writeIndent(&sb, 2, "enable: true")
				writeIndent(&sb, 1, "}")
			}
			if cfg.Optimization.ExecutionAccelerators != nil {
				writeIndent(&sb, 1, "execution_accelerators {")
				for _, ea := range cfg.Optimization.ExecutionAccelerators.GpuExecutionAccelerator {
					writeIndent(&sb, 2, "gpu_execution_accelerator {")
					writeIndent(&sb, 3, "name: %q", ea.Name)
					for _, param := range ea.Parameters {
						writeIndent(&sb, 3, "parameters {")
						writeIndent(&sb, 4, "key: %q", param.Key)
						writeIndent(&sb, 4, "value {")
						writeIndent(&sb, 5, "string_value: %q", param.Value.StringValue)
						writeIndent(&sb, 4, "}")
						writeIndent(&sb, 3, "}")
					}
					writeIndent(&sb, 2, "}")
				}
				for _, ea := range cfg.Optimization.ExecutionAccelerators.CpuExecutionAccelerator {
					writeIndent(&sb, 2, "cpu_execution_accelerator {")
					writeIndent(&sb, 3, "name: %q", ea.Name)
					for _, param := range ea.Parameters {
						writeIndent(&sb, 3, "parameters {")
						writeIndent(&sb, 4, "key: %q", param.Key)
						writeIndent(&sb, 4, "value {")
						writeIndent(&sb, 5, "string_value: %q", param.Value.StringValue)
						writeIndent(&sb, 4, "}")
						writeIndent(&sb, 3, "}")
					}
					writeIndent(&sb, 2, "}")
				}
				writeIndent(&sb, 1, "}")
			}
			writeIndent(&sb, 0, "}")
		}
	}

	// Parameters
	for _, p := range cfg.Parameters {
		writeIndent(&sb, 0, "parameters {")
		writeIndent(&sb, 1, "key: %q", p.Key)
		writeIndent(&sb, 1, "value {")
		writeIndent(&sb, 2, "string_value: %q", p.Value.StringValue)
		writeIndent(&sb, 1, "}")
		writeIndent(&sb, 0, "}")
	}

	// Warmup
	for _, w := range cfg.Warmups {
		writeIndent(&sb, 0, "model_warmup {")
		writeIndent(&sb, 1, "name: %q", w.Name)
		writeIndent(&sb, 1, "batch_size: %d", w.BatchSize)
		if w.Count > 0 {
			writeIndent(&sb, 1, "count: %d", w.Count)
		}
		for _, in := range w.Inputs {
			writeIndent(&sb, 1, "inputs {")
			writeIndent(&sb, 2, "key: %q", in.Key)
			writeIndent(&sb, 2, "value {")
			writeIndent(&sb, 3, "data_type: %s", in.Value.DataType)
			writeIndent(&sb, 3, "dims: [ %s ]", formatDims(in.Value.Dims))
			if in.Value.ZeroData {
				writeIndent(&sb, 3, "zero_data: true")
			}
			if in.Value.RandomData {
				writeIndent(&sb, 3, "random_data: true")
			}
			if in.Value.InputDataFile != "" {
				writeIndent(&sb, 3, "input_data_file: %q", in.Value.InputDataFile)
			}
			writeIndent(&sb, 2, "}")
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	// Response Cache
	if cfg.ResponseCache != nil && cfg.ResponseCache.Enable {
		writeIndent(&sb, 0, "response_cache {")
		writeIndent(&sb, 1, "enable: true")
		writeIndent(&sb, 0, "}")
	}

	// Ensemble Scheduling
	if cfg.EnsembleScheduling != nil {
		writeIndent(&sb, 0, "ensemble_scheduling {")
		for _, step := range cfg.EnsembleScheduling.Steps {
			writeIndent(&sb, 1, "step {")
			writeIndent(&sb, 2, "model_name: %q", step.ModelName)
			writeIndent(&sb, 2, "model_version: %d", step.ModelVersion)
			for _, m := range step.InputMaps {
				writeIndent(&sb, 2, "input_map {")
				writeIndent(&sb, 3, "key: %q", m.Key)
				writeIndent(&sb, 3, "value: %q", m.Value)
				writeIndent(&sb, 2, "}")
			}
			for _, m := range step.OutputMaps {
				writeIndent(&sb, 2, "output_map {")
				writeIndent(&sb, 3, "key: %q", m.Key)
				writeIndent(&sb, 3, "value: %q", m.Value)
				writeIndent(&sb, 2, "}")
			}
			writeIndent(&sb, 1, "}")
		}
		writeIndent(&sb, 0, "}")
	}

	return sb.String()
}
