package model

type ModelConfig struct {
	Name                 string              `json:"name"`
	Platform             string              `json:"platform"`
	Backend              string              `json:"backend"`
	MaxBatchSize         int32               `json:"max_batch_size"`
	DefaultModelFilename string              `json:"default_model_filename"`
	Inputs               []ModelInput        `json:"input"`
	Outputs              []ModelOutput       `json:"output"`
	VersionPolicy        *VersionPolicy      `json:"version_policy"`
	InstanceGroups       []InstanceGroup     `json:"instance_group"`
	DynamicBatching      *DynamicBatching    `json:"dynamic_batching"`
	SequenceBatching     *SequenceBatching   `json:"sequence_batching"`
	Optimization         *Optimization       `json:"optimization"`
	Parameters           []Parameter         `json:"parameters"`
	Warmups              []ModelWarmup       `json:"model_warmup"`
	ResponseCache        *ResponseCache      `json:"response_cache"`
	EnsembleScheduling   *EnsembleScheduling `json:"ensemble_scheduling"`
}

type ModelInput struct {
	Name             string   `json:"name"`
	DataType         string   `json:"data_type"`
	Dims             []int64  `json:"dims"`
	Reshape          *Reshape `json:"reshape"`
	IsShapeTensor    bool     `json:"is_shape_tensor"`
	AllowRaggedBatch bool     `json:"allow_ragged_batch"`
	Optional         bool     `json:"optional"`
}

type ModelOutput struct {
	Name          string   `json:"name"`
	DataType      string   `json:"data_type"`
	Dims          []int64  `json:"dims"`
	Reshape       *Reshape `json:"reshape"`
	LabelFilename string   `json:"label_filename"`
}

type Reshape struct {
	Dims []int64 `json:"dims"`
}

type VersionPolicy struct {
	Latest   *VersionPolicyLatest   `json:"latest"`
	All      *VersionPolicyAll      `json:"all"`
	Specific *VersionPolicySpecific `json:"specific"`
}

type VersionPolicyLatest struct {
	NumVersions int32 `json:"num_versions"`
}

type VersionPolicyAll struct{}

type VersionPolicySpecific struct {
	Versions []int64 `json:"versions"`
}

type InstanceGroup struct {
	Count      int32    `json:"count"`
	Kind       string   `json:"kind"` // KIND_AUTO, KIND_CPU, KIND_GPU, KIND_MODEL
	Gpus       []int32  `json:"gpus"`
	HostPolicy string   `json:"host_policy"`
}

type DynamicBatching struct {
	PreferredBatchSize         []int32               `json:"preferred_batch_size"`
	MaxQueueDelayMicroseconds  int64                 `json:"max_queue_delay_microseconds"`
	PreserveOrdering           bool                  `json:"preserve_ordering"`
	PriorityLevels             int32                 `json:"priority_levels"`
	DefaultQueuePolicy         *QueuePolicy          `json:"default_queue_policy"`
	PriorityQueuePolicy        []PriorityQueuePolicy `json:"priority_queue_policy"`
}

type QueuePolicy struct {
	TimeoutMicroseconds int64  `json:"timeout_microseconds"`
	MaxQueueSize        int32  `json:"max_queue_size"`
	Action              string `json:"action"` // REJECT, DELAY
}

type PriorityQueuePolicy struct {
	Priority    int32        `json:"priority"`
	QueuePolicy *QueuePolicy `json:"queue_policy"`
}

type SequenceBatching struct {
	MaxSequenceIdleMicroseconds int64                  `json:"max_sequence_idle_microseconds"`
	Direct                      *DirectSequenceBatcher `json:"direct"`
	Oldest                      *OldestSequenceBatcher `json:"oldest"`
	ControlInputs               []ControlInput         `json:"control_input"`
	States                      []SequenceState        `json:"state"`
}

type DirectSequenceBatcher struct {
	MinimumSlotAllocation int32 `json:"minimum_slot_allocation"`
}

type OldestSequenceBatcher struct {
	MaxQueueDelayMicroseconds int64 `json:"max_queue_delay_microseconds"`
}

type ControlInput struct {
	Name     string                 `json:"name"`
	Controls []ControlInputRelation `json:"control"`
}

type ControlInputRelation struct {
	Kind       string    `json:"kind"` // CONTROL_SEQUENCE_START, CONTROL_SEQUENCE_READY, etc.
	Int32Value []int32   `json:"int32_value"`
	Fp32Value  []float32 `json:"fp32_value"`
}

type SequenceState struct {
	Name         string         `json:"name"`
	DataType     string         `json:"data_type"`
	Dims         []int64        `json:"dims"`
	InitialState []InitialState `json:"initial_state"`
}

type InitialState struct {
	DataType string  `json:"data_type"`
	Dims     []int64 `json:"dims"`
	ZeroData bool    `json:"zero_data"`
	Name     string  `json:"name"`
}

type Optimization struct {
	ExecutionAccelerators *ExecutionAccelerators `json:"execution_accelerators"`
	InputPinnedMemory     bool                   `json:"input_pinned_memory"`
	OutputPinnedMemory    bool                   `json:"output_pinned_memory"`
}

type ExecutionAccelerators struct {
	GpuExecutionAccelerator []ExecutionAccelerator `json:"gpu_execution_accelerator"`
	CpuExecutionAccelerator []ExecutionAccelerator `json:"cpu_execution_accelerator"`
}

type ExecutionAccelerator struct {
	Name       string      `json:"name"`
	Parameters []Parameter `json:"parameters"`
}

type Parameter struct {
	Key   string         `json:"key"`
	Value ParameterValue `json:"value"`
}

type ParameterValue struct {
	StringValue string `json:"string_value"`
}

type ModelWarmup struct {
	Name      string        `json:"name"`
	BatchSize int32         `json:"batch_size"`
	Inputs    []WarmupInput `json:"inputs"`
	Count     int32         `json:"count"`
}

type WarmupInput struct {
	Key   string           `json:"key"`
	Value WarmupInputValue `json:"value"`
}

type WarmupInputValue struct {
	DataType      string  `json:"data_type"`
	Dims          []int64 `json:"dims"`
	ZeroData      bool    `json:"zero_data"`
	RandomData    bool    `json:"random_data"`
	InputDataFile string  `json:"input_data_file"`
}

type ResponseCache struct {
	Enable bool `json:"enable"`
}

type EnsembleScheduling struct {
	Steps []EnsembleStep `json:"step"`
}

type EnsembleStep struct {
	ModelName    string      `json:"model_name"`
	ModelVersion int64       `json:"model_version"`
	InputMaps    []StringMap `json:"input_map"`
	OutputMaps   []StringMap `json:"output_map"`
}

type StringMap struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
