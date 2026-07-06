package templates

import (
	"triton-config-studio/internal/model"
)

type Template struct {
	Name        string
	Description string
	Config      *model.ModelConfig
}

var BuiltInTemplates = []Template{
	{
		Name:        "TensorRT Image Classification",
		Description: "A standard configuration for TensorRT plan models performing image classification.",
		Config: &model.ModelConfig{
			Name:         "tensorrt_resnet",
			Platform:     "tensorrt_plan",
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
			InstanceGroups: []model.InstanceGroup{
				{
					Count: 1,
					Kind:  "KIND_GPU",
					Gpus:  []int32{0},
				},
			},
		},
	},
	{
		Name:        "PyTorch Computer Vision",
		Description: "Configuration for PyTorch LibTorch models with CPU/GPU serving.",
		Config: &model.ModelConfig{
			Name:         "pytorch_resnet",
			Backend:      "pytorch",
			MaxBatchSize: 16,
			Inputs: []model.ModelInput{
				{
					Name:     "INPUT__0",
					DataType: "TYPE_FP32",
					Dims:     []int64{3, 256, 256},
				},
			},
			Outputs: []model.ModelOutput{
				{
					Name:     "OUTPUT__0",
					DataType: "TYPE_FP32",
					Dims:     []int64{10},
				},
			},
			InstanceGroups: []model.InstanceGroup{
				{
					Count: 1,
					Kind:  "KIND_AUTO",
				},
			},
		},
	},
	{
		Name:        "ONNX NLP Transformer",
		Description: "Transformer configuration with dynamic sequence lengths and dynamic batching.",
		Config: &model.ModelConfig{
			Name:         "onnx_transformer",
			Platform:     "onnxruntime_onnx",
			MaxBatchSize: 32,
			Inputs: []model.ModelInput{
				{
					Name:     "input_ids",
					DataType: "TYPE_INT64",
					Dims:     []int64{-1},
				},
				{
					Name:     "attention_mask",
					DataType: "TYPE_INT64",
					Dims:     []int64{-1},
				},
			},
			Outputs: []model.ModelOutput{
				{
					Name:     "logits",
					DataType: "TYPE_FP32",
					Dims:     []int64{-1, 768},
				},
			},
			DynamicBatching: &model.DynamicBatching{
				PreferredBatchSize:        []int32{4, 8, 16, 32},
				MaxQueueDelayMicroseconds: 10000, // 10ms
			},
		},
	},
	{
		Name:        "Python LLM Service",
		Description: "A Python backend configuration optimized for serving LLM pipelines.",
		Config: &model.ModelConfig{
			Name:         "python_llm",
			Backend:      "python",
			MaxBatchSize: 128,
			Inputs: []model.ModelInput{
				{
					Name:     "text_input",
					DataType: "TYPE_STRING",
					Dims:     []int64{1},
				},
			},
			Outputs: []model.ModelOutput{
				{
					Name:     "text_output",
					DataType: "TYPE_STRING",
					Dims:     []int64{1},
				},
			},
			Parameters: []model.Parameter{
				{
					Key: "tokenizer_dir",
					Value: model.ParameterValue{
						StringValue: "./tokenizer",
					},
				},
			},
		},
	},
}
