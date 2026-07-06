package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"triton-config-studio/internal/fileio"
	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/state"
	"triton-config-studio/internal/templates"
	"triton-config-studio/internal/validator"
)

type EditorUI struct {
	window         fyne.Window
	state          *state.AppState
	sidebarList    *widget.List
	sections       []string
	activeSection  string
	rightContainer *fyne.Container
	statusLabel    *widget.Label
	valButton      *widget.Button
	valErrors      []string
}

func NewEditorUI(win fyne.Window, s *state.AppState) *EditorUI {
	ui := &EditorUI{
		window: win,
		state:  s,
		sections: []string{
			"General",
			"Inputs",
			"Outputs",
			"Version Policy",
			"Instance Groups",
			"Dynamic Batching",
			"Sequence Batching",
			"Optimization",
			"Parameters",
			"Warmup",
			"Response Cache",
			"Ensemble",
		},
		activeSection:  "General",
		rightContainer: container.NewMax(),
	}

	ui.state.RegisterListener(ui.onStateChanged)
	return ui
}

func (e *EditorUI) Build() fyne.CanvasObject {
	// 1. Sidebar List
	e.sidebarList = widget.NewList(
		func() int { return len(e.sections) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Section Name")
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(e.sections[id])
		},
	)
	e.sidebarList.OnSelected = func(id widget.ListItemID) {
		e.activeSection = e.sections[id]
		e.reloadActiveSection()
	}

	// 2. Toolbar Elements
	newBtn := widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), e.onNew)
	openBtn := widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), e.onOpen)
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), e.onSave)
	saveAsBtn := widget.NewButtonWithIcon("Save As", theme.DocumentSaveIcon(), e.onSaveAs)
	undoBtn := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), e.onUndo)
	redoBtn := widget.NewButtonWithIcon("", theme.ContentRedoIcon(), e.onRedo)
	genBtn := widget.NewButtonWithIcon("Generate", theme.ConfirmIcon(), e.onGenerate)

	// Templates selection dropdown
	var templateNames []string
	templateNames = append(templateNames, "Load Template...")
	for _, t := range templates.BuiltInTemplates {
		templateNames = append(templateNames, t.Name)
	}
	templateSelect := widget.NewSelect(templateNames, func(selected string) {
		if selected == "Load Template..." || selected == "" {
			return
		}
		for _, t := range templates.BuiltInTemplates {
			if t.Name == selected {
				e.state.SetConfig(t.Config)
				e.state.SetFilePath("")
				e.state.SetDirty(true)
				e.reloadActiveSection()
				dialog.ShowInformation("Template Loaded", fmt.Sprintf("Loaded %s template configuration.", t.Name), e.window)
				break
			}
		}
	})
	templateSelect.SetSelected("Load Template...")

	toolbar := container.NewHBox(
		newBtn, openBtn, saveBtn, saveAsBtn,
		widget.NewSeparator(),
		undoBtn, redoBtn,
		widget.NewSeparator(),
		genBtn,
		widget.NewSeparator(),
		templateSelect,
	)

	// 3. Right Pane Form Content
	e.reloadActiveSection()

	// 4. Split Layout
	split := container.NewHSplit(e.sidebarList, e.rightContainer)
	split.Offset = 0.2

	// 5. Status Bar Elements
	e.statusLabel = widget.NewLabel("File: New config.pbtxt")
	e.valButton = widget.NewButtonWithIcon("Validation: Checking...", theme.InfoIcon(), e.showValidationErrors)

	statusBar := container.NewBorder(
		nil, nil,
		e.statusLabel,
		e.valButton,
	)

	// Main Layout
	mainLayout := container.NewBorder(
		toolbar,
		statusBar,
		nil, nil,
		split,
	)

	// Trigger initial validation
	e.onStateChanged()

	return mainLayout
}

func (e *EditorUI) reloadActiveSection() {
	e.rightContainer.Objects = nil

	var form fyne.CanvasObject
	onModify := func() {
		// Callback for real-time validation refresh
		e.validateCurrentConfig()
	}

	onRebuild := func() {
		e.reloadActiveSection()
	}

	switch e.activeSection {
	case "General":
		form = buildGeneralForm(e.state, onModify)
	case "Inputs":
		form = buildInputsForm(e.state, onModify, onRebuild)
	case "Outputs":
		form = buildOutputsForm(e.state, onModify, onRebuild)
	case "Version Policy":
		form = buildVersionPolicyForm(e.state, onModify)
	case "Instance Groups":
		form = buildInstanceGroupsForm(e.state, onModify, onRebuild)
	case "Dynamic Batching":
		form = buildDynamicBatchingForm(e.state, onModify)
	case "Sequence Batching":
		form = buildSequenceBatchingForm(e.state, onModify)
	case "Optimization":
		form = buildOptimizationForm(e.state, onModify)
	case "Parameters":
		form = buildParametersForm(e.state, onModify, onRebuild)
	case "Warmup":
		form = buildWarmupForm(e.state, onModify, onRebuild)
	case "Response Cache":
		form = buildResponseCacheForm(e.state, onModify)
	case "Ensemble":
		form = buildEnsembleForm(e.state, onModify)
	default:
		form = widget.NewLabel("Select a configuration section from the sidebar.")
	}

	// Wrap in a ScrollContainer so it behaves nicely
	scroll := container.NewVScroll(form)
	e.rightContainer.Add(scroll)
	e.rightContainer.Refresh()
}

func (e *EditorUI) validateCurrentConfig() {
	cfg := e.state.GetConfig()
	e.valErrors = validator.Validate(cfg)
	
	// Append UI parsing/input validation errors
	uiErrs := e.state.GetUIErrors()
	for _, err := range uiErrs {
		e.valErrors = append(e.valErrors, err)
	}
	if len(e.valErrors) == 0 {
		e.valButton.SetText("✓ Configuration Valid")
		e.valButton.Importance = widget.SuccessImportance
		e.valButton.Icon = theme.ConfirmIcon()
	} else {
		e.valButton.SetText(fmt.Sprintf("✗ %d Validation Issues", len(e.valErrors)))
		e.valButton.Importance = widget.DangerImportance
		e.valButton.Icon = theme.WarningIcon()
	}
	e.valButton.Refresh()
}

func (e *EditorUI) showValidationErrors() {
	if len(e.valErrors) == 0 {
		dialog.ShowInformation("Validation Status", "No issues found! Configuration is fully valid.", e.window)
		return
	}

	var sb strings.Builder
	for _, err := range e.valErrors {
		sb.WriteString("- " + err + "\n")
	}

	scrollContent := container.NewVScroll(widget.NewLabel(sb.String()))
	scrollContent.SetMinSize(fyne.NewSize(500, 300))

	d := dialog.NewCustom("Validation Errors", "Close", scrollContent, e.window)
	d.Show()
}

func (e *EditorUI) onStateChanged() {
	// Update File Path and Dirty State
	path := e.state.GetFilePath()
	if path == "" {
		path = "Untitled.pbtxt"
	}
	dirtyStr := ""
	if e.state.IsDirty() {
		dirtyStr = " *"
	}
	e.statusLabel.SetText(fmt.Sprintf("File: %s%s", path, dirtyStr))
	e.statusLabel.Refresh()

	// Update Validation Status
	e.validateCurrentConfig()
}

func (e *EditorUI) onNew() {
	dialog.ShowConfirm("Discard Changes?", "Are you sure you want to create a new configuration? Unsaved changes will be lost.", func(discard bool) {
		if discard {
			s := state.NewAppState()
			e.state.SetConfig(s.GetConfig())
			e.state.SetFilePath("")
			e.state.SetDirty(false)
			e.reloadActiveSection()
		}
	}, e.window)
}

func (e *EditorUI) onOpen() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		path := reader.URI().Path()
		cfg, err := fileio.LoadConfig(path)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error loading pbtxt file: %w", err), e.window)
			return
		}

		e.state.SetConfig(cfg)
		e.state.SetFilePath(path)
		e.state.SetDirty(false)
		e.reloadActiveSection()
	}, e.window)

	// Filter for config.pbtxt or .pbtxt
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pbtxt"}))
	fd.Show()
}

func (e *EditorUI) onSave() {
	path := e.state.GetFilePath()
	if path == "" {
		e.onSaveAs()
		return
	}

	err := fileio.SaveConfig(path, e.state.GetConfig())
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	e.state.SetDirty(false)
	dialog.ShowInformation("File Saved", "Configuration saved successfully.", e.window)
}

func (e *EditorUI) onSaveAs() {
	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		path := writer.URI().Path()
		err = fileio.SaveConfig(path, e.state.GetConfig())
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}

		e.state.SetFilePath(path)
		e.state.SetDirty(false)
		dialog.ShowInformation("File Saved", "Configuration saved successfully.", e.window)
	}, e.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pbtxt"}))
	fd.SetFileName("config.pbtxt")
	fd.Show()
}

func (e *EditorUI) onUndo() {
	if e.state.CanUndo() {
		e.state.Undo()
		e.reloadActiveSection()
	}
}

func (e *EditorUI) onRedo() {
	if e.state.CanRedo() {
		e.state.Redo()
		e.reloadActiveSection()
	}
}

func (e *EditorUI) onGenerate() {
	content := generator.Generate(e.state.GetConfig())

	// Build a popover to show and let user copy to clipboard
	textEntry := widget.NewMultiLineEntry()
	textEntry.SetText(content)
	textEntryWrapper := container.NewGridWrap(fyne.NewSize(600, 400), textEntry)

	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		e.window.Clipboard().SetContent(content)
		dialog.ShowInformation("Copied", "Generated protobuf text copied to clipboard.", e.window)
	})

	contentBox := container.NewBorder(nil, copyBtn, nil, nil, textEntryWrapper)

	d := dialog.NewCustom("Generated config.pbtxt", "Close", contentBox, e.window)
	d.Show()
}
