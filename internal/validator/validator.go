package validator

import (
	"fmt"
	"regexp"

	"triton-config-studio/internal/model"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Validate checks a ModelConfig and returns a list of error/warning strings.
func Validate(cfg *model.ModelConfig) []string {
	var errs []string

	// 1. Model Name
	if cfg.Name == "" {
		errs = append(errs, "Error: Model name cannot be empty.")
	} else if !nameRegex.MatchString(cfg.Name) {
		errs = append(errs, fmt.Sprintf("Error: Model name %q is invalid. Only alphanumeric characters, underscores, and hyphens are allowed.", cfg.Name))
	}

	// 2. Platform / Backend
	if cfg.Platform == "" && cfg.Backend == "" {
		errs = append(errs, "Error: Either 'platform' or 'backend' must be specified.")
	}

	// 3. Max Batch Size
	if cfg.MaxBatchSize < 0 {
		errs = append(errs, fmt.Sprintf("Error: max_batch_size must be >= 0 (got %d).", cfg.MaxBatchSize))
	}

	// Track tensor names for uniqueness
	inputNames := make(map[string]bool)
	outputNames := make(map[string]bool)

	// 4. Inputs
	if len(cfg.Inputs) == 0 && cfg.Platform != "ensemble" && cfg.Backend != "ensemble" {
		errs = append(errs, "Warning: No inputs are defined. Most non-ensemble models require at least one input.")
	}

	for i, in := range cfg.Inputs {
		prefix := fmt.Sprintf("Input %d (%s)", i, in.Name)
		if in.Name == "" {
			errs = append(errs, fmt.Sprintf("Error: Input %d has an empty name.", i))
			continue
		}
		if inputNames[in.Name] {
			errs = append(errs, fmt.Sprintf("Error: Duplicate input name %q.", in.Name))
		}
		inputNames[in.Name] = true

		if in.DataType == "" {
			errs = append(errs, fmt.Sprintf("Error: %s has no data_type specified.", prefix))
		}

		if len(in.Dims) == 0 {
			errs = append(errs, fmt.Sprintf("Error: %s must have at least one dimension in 'dims'.", prefix))
		} else {
			for idx, d := range in.Dims {
				if d == 0 {
					errs = append(errs, fmt.Sprintf("Error: %s dimension %d is 0. Dimensions must be positive or -1 for dynamic dims.", prefix, idx))
				} else if d < -1 {
					errs = append(errs, fmt.Sprintf("Error: %s dimension %d is invalid (%d). Only positive integers and -1 are allowed.", prefix, idx, d))
				}
			}
		}

		if in.Reshape != nil {
			if len(in.Reshape.Dims) == 0 {
				errs = append(errs, fmt.Sprintf("Error: %s reshape is specified but contains no dimensions.", prefix))
			} else {
				for idx, d := range in.Reshape.Dims {
					if d == 0 || d < -1 {
						errs = append(errs, fmt.Sprintf("Error: %s reshape dimension %d is invalid (%d).", prefix, idx, d))
					}
				}
			}
		}
	}

	// 5. Outputs
	if len(cfg.Outputs) == 0 && cfg.Platform != "ensemble" && cfg.Backend != "ensemble" {
		errs = append(errs, "Warning: No outputs are defined. Most non-ensemble models require at least one output.")
	}

	for i, out := range cfg.Outputs {
		prefix := fmt.Sprintf("Output %d (%s)", i, out.Name)
		if out.Name == "" {
			errs = append(errs, fmt.Sprintf("Error: Output %d has an empty name.", i))
			continue
		}
		if outputNames[out.Name] {
			errs = append(errs, fmt.Sprintf("Error: Duplicate output name %q.", out.Name))
		}
		if inputNames[out.Name] {
			errs = append(errs, fmt.Sprintf("Error: Tensor name %q is used both as input and output.", out.Name))
		}
		outputNames[out.Name] = true

		if out.DataType == "" {
			errs = append(errs, fmt.Sprintf("Error: %s has no data_type specified.", prefix))
		}

		if len(out.Dims) == 0 {
			errs = append(errs, fmt.Sprintf("Error: %s must have at least one dimension in 'dims'.", prefix))
		} else {
			for idx, d := range out.Dims {
				if d == 0 {
					errs = append(errs, fmt.Sprintf("Error: %s dimension %d is 0. Dimensions must be positive or -1 for dynamic dims.", prefix, idx))
				} else if d < -1 {
					errs = append(errs, fmt.Sprintf("Error: %s dimension %d is invalid (%d). Only positive integers and -1 are allowed.", prefix, idx, d))
				}
			}
		}

		if out.Reshape != nil {
			if len(out.Reshape.Dims) == 0 {
				errs = append(errs, fmt.Sprintf("Error: %s reshape is specified but contains no dimensions.", prefix))
			} else {
				for idx, d := range out.Reshape.Dims {
					if d == 0 || d < -1 {
						errs = append(errs, fmt.Sprintf("Error: %s reshape dimension %d is invalid (%d).", prefix, idx, d))
					}
				}
			}
		}
	}

	// 6. Instance Groups
	for idx, grp := range cfg.InstanceGroups {
		prefix := fmt.Sprintf("Instance Group %d", idx)
		if grp.Count < 0 {
			errs = append(errs, fmt.Sprintf("Error: %s has invalid count %d. Count must be >= 0.", prefix, grp.Count))
		}
		if grp.Kind != "" {
			switch grp.Kind {
			case "KIND_AUTO", "KIND_CPU", "KIND_GPU", "KIND_MODEL":
				// valid
			default:
				errs = append(errs, fmt.Sprintf("Error: %s has invalid kind %q. Allowed kinds are KIND_AUTO, KIND_CPU, KIND_GPU, KIND_MODEL.", prefix, grp.Kind))
			}
		}
		for _, gpuId := range grp.Gpus {
			if gpuId < 0 {
				errs = append(errs, fmt.Sprintf("Error: %s has invalid GPU ID %d. GPU IDs must be >= 0.", prefix, gpuId))
			}
		}
	}

	// 7. Dynamic Batching
	if cfg.DynamicBatching != nil {
		for idx, pbSize := range cfg.DynamicBatching.PreferredBatchSize {
			if pbSize <= 0 {
				errs = append(errs, fmt.Sprintf("Error: Dynamic Batching preferred batch size %d at index %d is invalid. Must be > 0.", pbSize, idx))
			}
		}
		if cfg.DynamicBatching.MaxQueueDelayMicroseconds < 0 {
			errs = append(errs, fmt.Sprintf("Error: Dynamic Batching max_queue_delay_microseconds must be >= 0 (got %d).", cfg.DynamicBatching.MaxQueueDelayMicroseconds))
		}
		if cfg.DynamicBatching.PriorityLevels < 0 {
			errs = append(errs, fmt.Sprintf("Error: Dynamic Batching priority_levels must be >= 0 (got %d).", cfg.DynamicBatching.PriorityLevels))
		}
	}

	// 8. Parameters
	paramKeys := make(map[string]bool)
	for i, param := range cfg.Parameters {
		if param.Key == "" {
			errs = append(errs, fmt.Sprintf("Error: Parameter %d has an empty key.", i))
		} else {
			if paramKeys[param.Key] {
				errs = append(errs, fmt.Sprintf("Error: Duplicate parameter key %q.", param.Key))
			}
			paramKeys[param.Key] = true
		}
	}

	// 9. Warmup Samples
	for i, w := range cfg.Warmups {
		prefix := fmt.Sprintf("Warmup %d (%s)", i, w.Name)
		if w.Name == "" {
			errs = append(errs, fmt.Sprintf("Error: Warmup %d has an empty name.", i))
		}
		if w.BatchSize <= 0 {
			errs = append(errs, fmt.Sprintf("Error: %s batch_size must be > 0 (got %d).", prefix, w.BatchSize))
		}
		if w.Count < 0 {
			errs = append(errs, fmt.Sprintf("Error: %s count must be >= 0 (got %d).", prefix, w.Count))
		}
		for _, in := range w.Inputs {
			if in.Key == "" {
				errs = append(errs, fmt.Sprintf("Error: %s has a warmup input with an empty key.", prefix))
				continue
			}
			if !inputNames[in.Key] {
				errs = append(errs, fmt.Sprintf("Warning: %s warmup input %q does not match any model input name.", prefix, in.Key))
			}
			if in.Value.DataType == "" {
				errs = append(errs, fmt.Sprintf("Error: %s warmup input %q has no data_type specified.", prefix, in.Key))
			}
			if len(in.Value.Dims) == 0 {
				errs = append(errs, fmt.Sprintf("Error: %s warmup input %q must specify 'dims'.", prefix, in.Key))
			}
		}
	}

	return errs
}
