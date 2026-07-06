package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"triton-config-studio/internal/model"
	"triton-config-studio/internal/state"
)

var dataTypes = []string{
	"TYPE_INVALID", "TYPE_BOOL",
	"TYPE_UINT8", "TYPE_UINT16", "TYPE_UINT32", "TYPE_UINT64",
	"TYPE_INT8", "TYPE_INT16", "TYPE_INT32", "TYPE_INT64",
	"TYPE_FP16", "TYPE_FP32", "TYPE_FP64", "TYPE_STRING",
}

var platforms = []string{
	"", "tensorrt_plan", "onnxruntime_onnx", "tensorflow_graphdef",
	"tensorflow_savedmodel", "pytorch_libtorch", "ensemble",
}

var backends = []string{
	"", "tensorrt", "onnxruntime", "tensorflow", "pytorch", "python", "fil",
}

var instanceKinds = []string{
	"KIND_AUTO", "KIND_CPU", "KIND_GPU", "KIND_MODEL",
}

// Helper to parse dimensions like "[1, 28, 28]" or "1,28,28" or "1 28 28"
func parseDims(s string) []int64 {
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, ",", " ")
	fields := strings.Fields(s)
	var d []int64
	for _, f := range fields {
		val, err := strconv.ParseInt(f, 10, 64)
		if err == nil {
			d = append(d, val)
		}
	}
	return d
}

// Helper to parse GPU IDs like "0, 1"
func parseGpus(s string) []int32 {
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, ",", " ")
	fields := strings.Fields(s)
	var g []int32
	for _, f := range fields {
		val, err := strconv.ParseInt(f, 10, 32)
		if err == nil {
			g = append(g, int32(val))
		}
	}
	return g
}

func formatInt64Slice(slice []int64) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, ", ")
}

func formatInt32Slice(slice []int32) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, fmt.Sprintf("%d", v))
	}
	return strings.Join(parts, ", ")
}

// General settings form
func buildGeneralForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	nameEntry := widget.NewEntry()
	nameEntry.SetText(cfg.Name)
	nameEntry.OnChanged = func(val string) {
		cfg.Name = val
		s.SetDirty(true)
		onModify()
	}

	platformEntry := widget.NewSelectEntry(platforms)
	platformEntry.SetText(cfg.Platform)
	platformEntry.OnChanged = func(val string) {
		cfg.Platform = val
		s.SetDirty(true)
		onModify()
	}

	backendEntry := widget.NewSelectEntry(backends)
	backendEntry.SetText(cfg.Backend)
	backendEntry.OnChanged = func(val string) {
		cfg.Backend = val
		s.SetDirty(true)
		onModify()
	}

	maxBatchEntry := widget.NewEntry()
	maxBatchEntry.SetText(fmt.Sprintf("%d", cfg.MaxBatchSize))
	maxBatchEntry.OnChanged = func(val string) {
		if batchSize, err := strconv.ParseInt(val, 10, 32); err == nil {
			cfg.MaxBatchSize = int32(batchSize)
			s.SetDirty(true)
			onModify()
		}
	}

	defaultModelEntry := widget.NewEntry()
	defaultModelEntry.SetText(cfg.DefaultModelFilename)
	defaultModelEntry.OnChanged = func(val string) {
		cfg.DefaultModelFilename = val
		s.SetDirty(true)
		onModify()
	}

	form := widget.NewForm(
		widget.NewFormItem("Model Name", nameEntry),
		widget.NewFormItem("Platform", platformEntry),
		widget.NewFormItem("Backend", backendEntry),
		widget.NewFormItem("Max Batch Size", maxBatchEntry),
		widget.NewFormItem("Default Model Filename", defaultModelEntry),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("General Model Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
	)
}

// Inputs settings form
func buildInputsForm(s *state.AppState, onRebuild func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var selectedIndex = -1
	var rightContainer *fyne.Container
	var listWidget *widget.List

	// Rebuild right pane details
	showDetails := func(idx int) {
		rightContainer.Objects = nil
		if idx < 0 || idx >= len(cfg.Inputs) {
			rightContainer.Add(widget.NewLabel("Select an input tensor from the list to configure."))
			rightContainer.Refresh()
			return
		}

		in := &cfg.Inputs[idx]

		nameEntry := widget.NewEntry()
		nameEntry.SetText(in.Name)
		nameEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			in.Name = val
			listWidget.Refresh()
		}

		typeSelect := widget.NewSelect(dataTypes, func(val string) {
			s.SaveSnapshot()
			in.DataType = val
		})
		typeSelect.SetSelected(in.DataType)

		dimsEntry := widget.NewEntry()
		dimsEntry.SetText(formatInt64Slice(in.Dims))
		dimsEntry.SetPlaceHolder("e.g. 3, 224, 224 or -1, 768")
		dimsEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			in.Dims = parseDims(val)
		}

		reshapeEntry := widget.NewEntry()
		if in.Reshape != nil {
			reshapeEntry.SetText(formatInt64Slice(in.Reshape.Dims))
		}
		reshapeEntry.SetPlaceHolder("Leave blank if no reshape needed")
		reshapeEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			dims := parseDims(val)
			if len(dims) > 0 {
				in.Reshape = &model.Reshape{Dims: dims}
			} else {
				in.Reshape = nil
			}
		}

		optionalCheck := widget.NewCheck("Optional Input", func(checked bool) {
			s.SaveSnapshot()
			in.Optional = checked
		})
		optionalCheck.Checked = in.Optional

		raggedCheck := widget.NewCheck("Allow Ragged Batch", func(checked bool) {
			s.SaveSnapshot()
			in.AllowRaggedBatch = checked
		})
		raggedCheck.Checked = in.AllowRaggedBatch

		shapeCheck := widget.NewCheck("Is Shape Tensor", func(checked bool) {
			s.SaveSnapshot()
			in.IsShapeTensor = checked
		})
		shapeCheck.Checked = in.IsShapeTensor

		form := widget.NewForm(
			widget.NewFormItem("Tensor Name", nameEntry),
			widget.NewFormItem("Data Type", typeSelect),
			widget.NewFormItem("Dimensions", dimsEntry),
			widget.NewFormItem("Reshape Dimensions", reshapeEntry),
			widget.NewFormItem("", optionalCheck),
			widget.NewFormItem("", raggedCheck),
			widget.NewFormItem("", shapeCheck),
		)

		rightContainer.Add(container.NewVBox(
			widget.NewLabelWithStyle("Edit Input Tensor Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
		))
		rightContainer.Refresh()
	}

	// Left List
	listWidget = widget.NewList(
		func() int {
			return len(cfg.Inputs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Input Tensor")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			name := cfg.Inputs[i].Name
			if name == "" {
				name = "<unnamed>"
			}
			label.SetText(name)
		},
	)
	listWidget.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		showDetails(id)
	}

	// Add/Remove buttons
	addBtn := widget.NewButton("Add Input", func() {
		s.SaveSnapshot()
		cfg.Inputs = append(cfg.Inputs, model.ModelInput{
			Name:     fmt.Sprintf("input_%d", len(cfg.Inputs)),
			DataType: "TYPE_FP32",
			Dims:     []int64{-1},
		})
		listWidget.Refresh()
		listWidget.Select(len(cfg.Inputs) - 1)
	})

	removeBtn := widget.NewButton("Remove Input", func() {
		if selectedIndex >= 0 && selectedIndex < len(cfg.Inputs) {
			s.SaveSnapshot()
			cfg.Inputs = append(cfg.Inputs[:selectedIndex], cfg.Inputs[selectedIndex+1:]...)
			selectedIndex = -1
			listWidget.Refresh()
			listWidget.UnselectAll()
			showDetails(-1)
		}
	})

	leftPane := container.NewBorder(
		nil,
		container.NewHBox(addBtn, removeBtn),
		nil,
		nil,
		listWidget,
	)

	rightContainer = container.NewVBox()
	showDetails(-1)

	split := container.NewHSplit(leftPane, container.NewVScroll(rightContainer))
	split.Offset = 0.3

	return container.New(layout.NewMaxLayout(), split)
}

// Outputs settings form
func buildOutputsForm(s *state.AppState, onRebuild func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var selectedIndex = -1
	var rightContainer *fyne.Container
	var listWidget *widget.List

	showDetails := func(idx int) {
		rightContainer.Objects = nil
		if idx < 0 || idx >= len(cfg.Outputs) {
			rightContainer.Add(widget.NewLabel("Select an output tensor from the list to configure."))
			rightContainer.Refresh()
			return
		}

		out := &cfg.Outputs[idx]

		nameEntry := widget.NewEntry()
		nameEntry.SetText(out.Name)
		nameEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			out.Name = val
			listWidget.Refresh()
		}

		typeSelect := widget.NewSelect(dataTypes, func(val string) {
			s.SaveSnapshot()
			out.DataType = val
		})
		typeSelect.SetSelected(out.DataType)

		dimsEntry := widget.NewEntry()
		dimsEntry.SetText(formatInt64Slice(out.Dims))
		dimsEntry.SetPlaceHolder("e.g. 1000 or -1, 768")
		dimsEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			out.Dims = parseDims(val)
		}

		reshapeEntry := widget.NewEntry()
		if out.Reshape != nil {
			reshapeEntry.SetText(formatInt64Slice(out.Reshape.Dims))
		}
		reshapeEntry.SetPlaceHolder("Leave blank if no reshape needed")
		reshapeEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			dims := parseDims(val)
			if len(dims) > 0 {
				out.Reshape = &model.Reshape{Dims: dims}
			} else {
				out.Reshape = nil
			}
		}

		labelEntry := widget.NewEntry()
		labelEntry.SetText(out.LabelFilename)
		labelEntry.SetPlaceHolder("e.g. labels.txt")
		labelEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			out.LabelFilename = val
		}

		form := widget.NewForm(
			widget.NewFormItem("Tensor Name", nameEntry),
			widget.NewFormItem("Data Type", typeSelect),
			widget.NewFormItem("Dimensions", dimsEntry),
			widget.NewFormItem("Reshape Dimensions", reshapeEntry),
			widget.NewFormItem("Label Filename", labelEntry),
		)

		rightContainer.Add(container.NewVBox(
			widget.NewLabelWithStyle("Edit Output Tensor Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
		))
		rightContainer.Refresh()
	}

	listWidget = widget.NewList(
		func() int {
			return len(cfg.Outputs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Output Tensor")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			name := cfg.Outputs[i].Name
			if name == "" {
				name = "<unnamed>"
			}
			label.SetText(name)
		},
	)
	listWidget.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		showDetails(id)
	}

	addBtn := widget.NewButton("Add Output", func() {
		s.SaveSnapshot()
		cfg.Outputs = append(cfg.Outputs, model.ModelOutput{
			Name:     fmt.Sprintf("output_%d", len(cfg.Outputs)),
			DataType: "TYPE_FP32",
			Dims:     []int64{-1},
		})
		listWidget.Refresh()
		listWidget.Select(len(cfg.Outputs) - 1)
	})

	removeBtn := widget.NewButton("Remove Output", func() {
		if selectedIndex >= 0 && selectedIndex < len(cfg.Outputs) {
			s.SaveSnapshot()
			cfg.Outputs = append(cfg.Outputs[:selectedIndex], cfg.Outputs[selectedIndex+1:]...)
			selectedIndex = -1
			listWidget.Refresh()
			listWidget.UnselectAll()
			showDetails(-1)
		}
	})

	leftPane := container.NewBorder(
		nil,
		container.NewHBox(addBtn, removeBtn),
		nil,
		nil,
		listWidget,
	)

	rightContainer = container.NewVBox()
	showDetails(-1)

	split := container.NewHSplit(leftPane, container.NewVScroll(rightContainer))
	split.Offset = 0.3

	return container.New(layout.NewMaxLayout(), split)
}

// Version Policy form
func buildVersionPolicyForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	options := []string{"Latest", "All", "Specific"}
	var selectWidget *widget.Select
	var cardContainer *fyne.Container

	rebuildDetails := func(polType string) {
		cardContainer.Objects = nil
		s.SaveSnapshot()

		switch polType {
		case "Latest":
			cfg.VersionPolicy = &model.VersionPolicy{
				Latest: &model.VersionPolicyLatest{NumVersions: 1},
			}
			entry := widget.NewEntry()
			entry.SetText("1")
			entry.OnChanged = func(val string) {
				if num, err := strconv.Atoi(val); err == nil {
					cfg.VersionPolicy.Latest.NumVersions = int32(num)
					onModify()
				}
			}
			cardContainer.Add(widget.NewForm(widget.NewFormItem("Number of Latest Versions", entry)))

		case "All":
			cfg.VersionPolicy = &model.VersionPolicy{
				All: &model.VersionPolicyAll{},
			}
			cardContainer.Add(widget.NewLabel("All available versions of the model will be loaded by Triton."))

		case "Specific":
			cfg.VersionPolicy = &model.VersionPolicy{
				Specific: &model.VersionPolicySpecific{Versions: []int64{1}},
			}
			entry := widget.NewEntry()
			entry.SetText("1")
			entry.SetPlaceHolder("e.g. 1, 2, 5")
			entry.OnChanged = func(val string) {
				cfg.VersionPolicy.Specific.Versions = parseDims(val)
				onModify()
			}
			cardContainer.Add(widget.NewForm(widget.NewFormItem("Version Numbers", entry)))

		default:
			cfg.VersionPolicy = nil
		}
		cardContainer.Refresh()
	}

	currentPolicy := "None"
	if cfg.VersionPolicy != nil {
		if cfg.VersionPolicy.Latest != nil {
			currentPolicy = "Latest"
		} else if cfg.VersionPolicy.All != nil {
			currentPolicy = "All"
		} else if cfg.VersionPolicy.Specific != nil {
			currentPolicy = "Specific"
		}
	}

	selectWidget = widget.NewSelect(options, rebuildDetails)
	selectWidget.SetSelected(currentPolicy)

	cardContainer = container.NewVBox()
	if currentPolicy != "None" {
		rebuildDetails(currentPolicy)
	}

	return container.NewVBox(
		widget.NewLabelWithStyle("Version Policy Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewForm(widget.NewFormItem("Policy Mode", selectWidget)),
		cardContainer,
	)
}

// Instance Groups settings form
func buildInstanceGroupsForm(s *state.AppState, onRebuild func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var selectedIndex = -1
	var rightContainer *fyne.Container
	var listWidget *widget.List

	showDetails := func(idx int) {
		rightContainer.Objects = nil
		if idx < 0 || idx >= len(cfg.InstanceGroups) {
			rightContainer.Add(widget.NewLabel("Select an instance group from the list to configure."))
			rightContainer.Refresh()
			return
		}

		grp := &cfg.InstanceGroups[idx]

		countEntry := widget.NewEntry()
		countEntry.SetText(fmt.Sprintf("%d", grp.Count))
		countEntry.OnChanged = func(val string) {
			if count, err := strconv.Atoi(val); err == nil {
				s.SaveSnapshot()
				grp.Count = int32(count)
				listWidget.Refresh()
			}
		}

		kindSelect := widget.NewSelect(instanceKinds, func(val string) {
			s.SaveSnapshot()
			grp.Kind = val
			listWidget.Refresh()
		})
		kindSelect.SetSelected(grp.Kind)

		gpusEntry := widget.NewEntry()
		gpusEntry.SetText(formatInt32Slice(grp.Gpus))
		gpusEntry.SetPlaceHolder("e.g. 0, 1 (Leave empty for CPU or Autoselect)")
		gpusEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			grp.Gpus = parseGpus(val)
		}

		hostPolicyEntry := widget.NewEntry()
		hostPolicyEntry.SetText(grp.HostPolicy)
		hostPolicyEntry.SetPlaceHolder("Host policy name")
		hostPolicyEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			grp.HostPolicy = val
		}

		form := widget.NewForm(
			widget.NewFormItem("Instance Count", countEntry),
			widget.NewFormItem("Device Kind", kindSelect),
			widget.NewFormItem("GPU IDs", gpusEntry),
			widget.NewFormItem("Host Policy", hostPolicyEntry),
		)

		rightContainer.Add(container.NewVBox(
			widget.NewLabelWithStyle("Edit Instance Group Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
		))
		rightContainer.Refresh()
	}

	listWidget = widget.NewList(
		func() int {
			return len(cfg.InstanceGroups)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Instance Group")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			g := cfg.InstanceGroups[i]
			kind := g.Kind
			if kind == "" {
				kind = "KIND_AUTO"
			}
			label.SetText(fmt.Sprintf("Group %d (%s x%d)", i, kind, g.Count))
		},
	)
	listWidget.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		showDetails(id)
	}

	addBtn := widget.NewButton("Add Group", func() {
		s.SaveSnapshot()
		cfg.InstanceGroups = append(cfg.InstanceGroups, model.InstanceGroup{
			Count: 1,
			Kind:  "KIND_GPU",
			Gpus:  []int32{0},
		})
		listWidget.Refresh()
		listWidget.Select(len(cfg.InstanceGroups) - 1)
	})

	removeBtn := widget.NewButton("Remove Group", func() {
		if selectedIndex >= 0 && selectedIndex < len(cfg.InstanceGroups) {
			s.SaveSnapshot()
			cfg.InstanceGroups = append(cfg.InstanceGroups[:selectedIndex], cfg.InstanceGroups[selectedIndex+1:]...)
			selectedIndex = -1
			listWidget.Refresh()
			listWidget.UnselectAll()
			showDetails(-1)
		}
	})

	leftPane := container.NewBorder(
		nil,
		container.NewHBox(addBtn, removeBtn),
		nil,
		nil,
		listWidget,
	)

	rightContainer = container.NewVBox()
	showDetails(-1)

	split := container.NewHSplit(leftPane, container.NewVScroll(rightContainer))
	split.Offset = 0.3

	return container.New(layout.NewMaxLayout(), split)
}

// Dynamic Batching settings form
func buildDynamicBatchingForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var dynamicForm *fyne.Container
	var enableCheck *widget.Check

	rebuildFields := func(enabled bool) {
		dynamicForm.Objects = nil
		if !enabled {
			cfg.DynamicBatching = nil
			s.SetDirty(true)
			onModify()
			return
		}

		if cfg.DynamicBatching == nil {
			cfg.DynamicBatching = &model.DynamicBatching{
				PreferredBatchSize:        []int32{},
				MaxQueueDelayMicroseconds: 0,
			}
		}

		prefSizesEntry := widget.NewEntry()
		prefSizesEntry.SetText(formatInt32Slice(cfg.DynamicBatching.PreferredBatchSize))
		prefSizesEntry.SetPlaceHolder("e.g. 2, 4, 8, 16")
		prefSizesEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			cfg.DynamicBatching.PreferredBatchSize = parseGpus(val)
			onModify()
		}

		delayEntry := widget.NewEntry()
		delayEntry.SetText(fmt.Sprintf("%d", cfg.DynamicBatching.MaxQueueDelayMicroseconds))
		delayEntry.SetPlaceHolder("in microseconds, e.g. 5000")
		delayEntry.OnChanged = func(val string) {
			if delay, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.SaveSnapshot()
				cfg.DynamicBatching.MaxQueueDelayMicroseconds = delay
				onModify()
			}
		}

		preserveOrderCheck := widget.NewCheck("Preserve Order", func(checked bool) {
			s.SaveSnapshot()
			cfg.DynamicBatching.PreserveOrdering = checked
			onModify()
		})
		preserveOrderCheck.Checked = cfg.DynamicBatching.PreserveOrdering

		priorityLevelsEntry := widget.NewEntry()
		priorityLevelsEntry.SetText(fmt.Sprintf("%d", cfg.DynamicBatching.PriorityLevels))
		priorityLevelsEntry.OnChanged = func(val string) {
			if levels, err := strconv.Atoi(val); err == nil {
				s.SaveSnapshot()
				cfg.DynamicBatching.PriorityLevels = int32(levels)
				onModify()
			}
		}

		form := widget.NewForm(
			widget.NewFormItem("Preferred Batch Sizes", prefSizesEntry),
			widget.NewFormItem("Max Queue Delay (μs)", delayEntry),
			widget.NewFormItem("Ordering Policy", preserveOrderCheck),
			widget.NewFormItem("Priority Levels", priorityLevelsEntry),
		)

		dynamicForm.Add(form)
		dynamicForm.Refresh()
	}

	enabled := cfg.DynamicBatching != nil
	enableCheck = widget.NewCheck("Enable Dynamic Batching", rebuildFields)
	enableCheck.Checked = enabled

	dynamicForm = container.NewVBox()
	rebuildFields(enabled)

	return container.NewVBox(
		widget.NewLabelWithStyle("Dynamic Batching", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		enableCheck,
		dynamicForm,
	)
}

// Sequence Batching settings form
func buildSequenceBatchingForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var sequenceForm *fyne.Container
	var enableCheck *widget.Check

	rebuildFields := func(enabled bool) {
		sequenceForm.Objects = nil
		if !enabled {
			cfg.SequenceBatching = nil
			s.SetDirty(true)
			onModify()
			return
		}

		if cfg.SequenceBatching == nil {
			cfg.SequenceBatching = &model.SequenceBatching{
				MaxSequenceIdleMicroseconds: 60000000, // 60s default
			}
		}

		idleEntry := widget.NewEntry()
		idleEntry.SetText(fmt.Sprintf("%d", cfg.SequenceBatching.MaxSequenceIdleMicroseconds))
		idleEntry.OnChanged = func(val string) {
			if idle, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.SaveSnapshot()
				cfg.SequenceBatching.MaxSequenceIdleMicroseconds = idle
				onModify()
			}
		}

		form := widget.NewForm(
			widget.NewFormItem("Max Sequence Idle (μs)", idleEntry),
		)

		sequenceForm.Add(form)
		sequenceForm.Refresh()
	}

	enabled := cfg.SequenceBatching != nil
	enableCheck = widget.NewCheck("Enable Sequence Batching", rebuildFields)
	enableCheck.Checked = enabled

	sequenceForm = container.NewVBox()
	rebuildFields(enabled)

	return container.NewVBox(
		widget.NewLabelWithStyle("Sequence Batching", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		enableCheck,
		sequenceForm,
	)
}

// Optimization settings form
func buildOptimizationForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	if cfg.Optimization == nil {
		cfg.Optimization = &model.Optimization{}
	}

	pinnedInputCheck := widget.NewCheck("Input Pinned Memory", func(checked bool) {
		s.SaveSnapshot()
		cfg.Optimization.InputPinnedMemory = checked
		onModify()
	})
	pinnedInputCheck.Checked = cfg.Optimization.InputPinnedMemory

	pinnedOutputCheck := widget.NewCheck("Output Pinned Memory", func(checked bool) {
		s.SaveSnapshot()
		cfg.Optimization.OutputPinnedMemory = checked
		onModify()
	})
	pinnedOutputCheck.Checked = cfg.Optimization.OutputPinnedMemory

	form := widget.NewForm(
		widget.NewFormItem("Input Memory", pinnedInputCheck),
		widget.NewFormItem("Output Memory", pinnedOutputCheck),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Optimization Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
	)
}

// Parameters settings form
func buildParametersForm(s *state.AppState, onRebuild func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var selectedIndex = -1
	var rightContainer *fyne.Container
	var listWidget *widget.List

	showDetails := func(idx int) {
		rightContainer.Objects = nil
		if idx < 0 || idx >= len(cfg.Parameters) {
			rightContainer.Add(widget.NewLabel("Select a parameter to edit."))
			rightContainer.Refresh()
			return
		}

		p := &cfg.Parameters[idx]

		keyEntry := widget.NewEntry()
		keyEntry.SetText(p.Key)
		keyEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			p.Key = val
			listWidget.Refresh()
		}

		valueEntry := widget.NewEntry()
		valueEntry.SetText(p.Value.StringValue)
		valueEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			p.Value.StringValue = val
		}

		form := widget.NewForm(
			widget.NewFormItem("Parameter Key", keyEntry),
			widget.NewFormItem("Parameter Value", valueEntry),
		)

		rightContainer.Add(container.NewVBox(
			widget.NewLabelWithStyle("Edit Parameter details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
		))
		rightContainer.Refresh()
	}

	listWidget = widget.NewList(
		func() int {
			return len(cfg.Parameters)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Parameter")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			k := cfg.Parameters[i].Key
			if k == "" {
				k = "<empty>"
			}
			label.SetText(k)
		},
	)
	listWidget.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		showDetails(id)
	}

	addBtn := widget.NewButton("Add Parameter", func() {
		s.SaveSnapshot()
		cfg.Parameters = append(cfg.Parameters, model.Parameter{
			Key: fmt.Sprintf("param_%d", len(cfg.Parameters)),
			Value: model.ParameterValue{
				StringValue: "",
			},
		})
		listWidget.Refresh()
		listWidget.Select(len(cfg.Parameters) - 1)
	})

	removeBtn := widget.NewButton("Remove Parameter", func() {
		if selectedIndex >= 0 && selectedIndex < len(cfg.Parameters) {
			s.SaveSnapshot()
			cfg.Parameters = append(cfg.Parameters[:selectedIndex], cfg.Parameters[selectedIndex+1:]...)
			selectedIndex = -1
			listWidget.Refresh()
			listWidget.UnselectAll()
			showDetails(-1)
		}
	})

	leftPane := container.NewBorder(
		nil,
		container.NewHBox(addBtn, removeBtn),
		nil,
		nil,
		listWidget,
	)

	rightContainer = container.NewVBox()
	showDetails(-1)

	split := container.NewHSplit(leftPane, container.NewVScroll(rightContainer))
	split.Offset = 0.3

	return container.New(layout.NewMaxLayout(), split)
}

// Warmup settings form
func buildWarmupForm(s *state.AppState, onRebuild func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var selectedIndex = -1
	var rightContainer *fyne.Container
	var listWidget *widget.List

	showDetails := func(idx int) {
		rightContainer.Objects = nil
		if idx < 0 || idx >= len(cfg.Warmups) {
			rightContainer.Add(widget.NewLabel("Select a warmup sample to edit."))
			rightContainer.Refresh()
			return
		}

		w := &cfg.Warmups[idx]

		nameEntry := widget.NewEntry()
		nameEntry.SetText(w.Name)
		nameEntry.OnChanged = func(val string) {
			s.SaveSnapshot()
			w.Name = val
			listWidget.Refresh()
		}

		batchEntry := widget.NewEntry()
		batchEntry.SetText(fmt.Sprintf("%d", w.BatchSize))
		batchEntry.OnChanged = func(val string) {
			if b, err := strconv.Atoi(val); err == nil {
				s.SaveSnapshot()
				w.BatchSize = int32(b)
			}
		}

		countEntry := widget.NewEntry()
		countEntry.SetText(fmt.Sprintf("%d", w.Count))
		countEntry.OnChanged = func(val string) {
			if c, err := strconv.Atoi(val); err == nil {
				s.SaveSnapshot()
				w.Count = int32(c)
			}
		}

		form := widget.NewForm(
			widget.NewFormItem("Sample Name", nameEntry),
			widget.NewFormItem("Batch Size", batchEntry),
			widget.NewFormItem("Execution Count", countEntry),
		)

		rightContainer.Add(container.NewVBox(
			widget.NewLabelWithStyle("Edit Warmup Sample Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
			widget.NewLabel("Warmup Inputs can be defined in the pbtxt file directly."),
		))
		rightContainer.Refresh()
	}

	listWidget = widget.NewList(
		func() int {
			return len(cfg.Warmups)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Warmup Sample")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			name := cfg.Warmups[i].Name
			if name == "" {
				name = "<unnamed>"
			}
			label.SetText(name)
		},
	)
	listWidget.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
		showDetails(id)
	}

	addBtn := widget.NewButton("Add Warmup", func() {
		s.SaveSnapshot()
		cfg.Warmups = append(cfg.Warmups, model.ModelWarmup{
			Name:      fmt.Sprintf("warmup_sample_%d", len(cfg.Warmups)),
			BatchSize: 1,
			Count:     1,
		})
		listWidget.Refresh()
		listWidget.Select(len(cfg.Warmups) - 1)
	})

	removeBtn := widget.NewButton("Remove Warmup", func() {
		if selectedIndex >= 0 && selectedIndex < len(cfg.Warmups) {
			s.SaveSnapshot()
			cfg.Warmups = append(cfg.Warmups[:selectedIndex], cfg.Warmups[selectedIndex+1:]...)
			selectedIndex = -1
			listWidget.Refresh()
			listWidget.UnselectAll()
			showDetails(-1)
		}
	})

	leftPane := container.NewBorder(
		nil,
		container.NewHBox(addBtn, removeBtn),
		nil,
		nil,
		listWidget,
	)

	rightContainer = container.NewVBox()
	showDetails(-1)

	split := container.NewHSplit(leftPane, container.NewVScroll(rightContainer))
	split.Offset = 0.3

	return container.New(layout.NewMaxLayout(), split)
}

// Response Cache form
func buildResponseCacheForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	if cfg.ResponseCache == nil {
		cfg.ResponseCache = &model.ResponseCache{}
	}

	enableCheck := widget.NewCheck("Enable Response Cache", func(checked bool) {
		s.SaveSnapshot()
		cfg.ResponseCache.Enable = checked
		onModify()
	})
	enableCheck.Checked = cfg.ResponseCache.Enable

	form := widget.NewForm(
		widget.NewFormItem("Response Cache", enableCheck),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Response Cache Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
	)
}

// Ensemble form
func buildEnsembleForm(s *state.AppState, onModify func()) fyne.CanvasObject {
	cfg := s.GetConfig()

	var dynamicForm *fyne.Container
	var enableCheck *widget.Check

	rebuildFields := func(enabled bool) {
		dynamicForm.Objects = nil
		if !enabled {
			cfg.EnsembleScheduling = nil
			s.SetDirty(true)
			onModify()
			return
		}

		if cfg.EnsembleScheduling == nil {
			cfg.EnsembleScheduling = &model.EnsembleScheduling{
				Steps: []model.EnsembleStep{},
			}
		}

		dynamicForm.Add(widget.NewLabel("Ensemble Scheduling is enabled. Configure steps using manual text mode, or click dynamic elements to save."))
		dynamicForm.Refresh()
	}

	enabled := cfg.EnsembleScheduling != nil
	enableCheck = widget.NewCheck("Enable Ensemble Scheduling", rebuildFields)
	enableCheck.Checked = enabled

	dynamicForm = container.NewVBox()
	rebuildFields(enabled)

	return container.NewVBox(
		widget.NewLabelWithStyle("Ensemble Scheduling", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		enableCheck,
		dynamicForm,
	)
}
