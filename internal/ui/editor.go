package ui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"triton-config-studio/internal/fileio"
	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/model"
	"triton-config-studio/internal/state"
	"triton-config-studio/internal/templates"
	"triton-config-studio/internal/validator"
)

type EditorUI struct {
	window           fyne.Window
	state            *state.AppState
	sidebarList      *widget.List
	sections         []string
	activeSection    string
	rightContainer   *fyne.Container
	statusLabel      *widget.Label
	valButton        *widget.Button
	valErrors        []string

	// Live Preview Panel components
	previewEntry     *readOnlyEntry
	previewContainer *fyne.Container
	contentContainer *fyne.Container
	previewButton    *widget.Button
	previewVisible   bool
	mainSplit        *container.Split
	outerSplit       *container.Split

	// Toolbar undo/redo buttons (removed)

	// Selection tracking to keep active list element selected after Undo/Redo/Rebuild
	selectedInputsIdx         int
	selectedOutputsIdx        int
	selectedInstanceGroupsIdx int
	selectedParametersIdx     int
	selectedWarmupsIdx        int
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
			"Versions & Models",
			"Export Repository",
		},
		activeSection:             "General",
		rightContainer:            container.NewMax(),
		previewVisible:            true, // Default live preview panel to visible
		selectedInputsIdx:         -1,
		selectedOutputsIdx:        -1,
		selectedInstanceGroupsIdx: -1,
		selectedParametersIdx:     -1,
		selectedWarmupsIdx:        -1,
	}

	ui.state.RegisterListener(ui.onStateChanged)

	// Register Ctrl+S / Cmd+S shortcut for Save
	win.Canvas().AddShortcut(&desktop.CustomShortcut{
		KeyName:  fyne.KeyS,
		Modifier: fyne.KeyModifierShortcutDefault,
	}, func(shortcut fyne.Shortcut) {
		ui.onSave()
	})

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

	// 2. Live Preview Panel initialization
	e.previewEntry = newReadOnlyEntry()
	e.previewEntry.TextStyle = fyne.TextStyle{Monospace: true}

	copyBtn := widget.NewButtonWithIcon("Copy Configuration", theme.ContentCopyIcon(), func() {
		e.window.Clipboard().SetContent(e.previewEntry.Text)
		dialog.ShowInformation("Copied", "Configuration copied to clipboard.", e.window)
	})
	e.previewContainer = container.NewBorder(
		widget.NewLabelWithStyle("Live Preview (config.pbtxt)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		copyBtn,
		nil,
		nil,
		container.NewThemeOverride(e.previewEntry, &previewTheme{Theme: fyne.CurrentApp().Settings().Theme()}),
	)

	// 3. Toolbar Elements
	newBtn := widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), e.onNew)
	openBtn := widget.NewButtonWithIcon("Open Folder", theme.FolderOpenIcon(), e.onOpen)
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), e.onSave)

	// Togglable Preview button
	e.previewButton = widget.NewButtonWithIcon("Preview", theme.VisibilityIcon(), e.togglePreview)

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
				e.state.SetModelFolderPath("")
				e.state.SetDirty(true)
				e.reloadActiveSection()
				dialog.ShowInformation("Template Loaded", fmt.Sprintf("Loaded %s template configuration.", t.Name), e.window)
				break
			}
		}
	})
	templateSelect.SetSelected("Load Template...")

	toolbar := container.NewHBox(
		newBtn, openBtn, saveBtn,
		widget.NewSeparator(),
		e.previewButton,
		widget.NewSeparator(),
		templateSelect,
	)

	// 4. Right Pane Form Content (Active Section)
	e.reloadActiveSection()

	// 5. Layout Container setup
	e.contentContainer = container.NewMax()
	e.updateLayout()

	// 6. Status Bar Elements
	e.statusLabel = widget.NewLabel("Folder: No folder opened")
	e.valButton = widget.NewButtonWithIcon("Validation: Checking...", theme.InfoIcon(), e.showValidationErrors)

	statusBar := container.NewBorder(
		nil, nil,
		e.statusLabel,
		e.valButton,
	)

	// Main Layout assembly
	mainLayout := container.NewBorder(
		toolbar,
		statusBar,
		nil, nil,
		e.contentContainer,
	)

	// Trigger initial validation & preview text update
	e.onStateChanged()

	return mainLayout
}

func (e *EditorUI) updateLayout() {
	e.contentContainer.Objects = nil

	// Forms split contains: Sidebar + Form fields
	e.mainSplit = container.NewHSplit(e.sidebarList, e.rightContainer)
	e.mainSplit.Offset = 0.2

	if e.previewVisible {
		// Outer split contains: Sidebar-Form split + Preview Panel
		e.outerSplit = container.NewHSplit(e.mainSplit, e.previewContainer)
		e.outerSplit.Offset = 0.65
		e.contentContainer.Add(e.outerSplit)
	} else {
		e.contentContainer.Add(e.mainSplit)
	}

	e.contentContainer.Refresh()
}

func (e *EditorUI) togglePreview() {
	e.previewVisible = !e.previewVisible
	if e.previewVisible {
		e.previewButton.Icon = theme.VisibilityIcon()
		e.updatePreviewText()
	} else {
		e.previewButton.Icon = theme.VisibilityOffIcon()
	}
	e.previewButton.Refresh()
	e.updateLayout()
}

func (e *EditorUI) updatePreviewText() {
	if e.previewVisible {
		content := generator.Generate(e.state.GetConfig())
		e.previewEntry.SetText(content)
		e.previewEntry.Refresh()
	}
}

func (e *EditorUI) reloadActiveSection() {
	e.rightContainer.Objects = nil

	var form fyne.CanvasObject
	onModify := func() {
		e.state.SetDirty(true)
		// Callback for real-time validation refresh
		e.validateCurrentConfig()
		e.updatePreviewText()
	}

	onRebuild := func() {
		e.reloadActiveSection()
	}

	switch e.activeSection {
	case "General":
		form = buildGeneralForm(e.state, onModify)
	case "Inputs":
		form = buildInputsForm(e.state, &e.selectedInputsIdx, onModify, onRebuild)
	case "Outputs":
		form = buildOutputsForm(e.state, &e.selectedOutputsIdx, onModify, onRebuild)
	case "Version Policy":
		form = buildVersionPolicyForm(e.state, onModify)
	case "Instance Groups":
		form = buildInstanceGroupsForm(e.state, &e.selectedInstanceGroupsIdx, onModify, onRebuild)
	case "Dynamic Batching":
		form = buildDynamicBatchingForm(e.state, onModify)
	case "Sequence Batching":
		form = buildSequenceBatchingForm(e.state, onModify)
	case "Optimization":
		form = buildOptimizationForm(e.state, onModify)
	case "Parameters":
		form = buildParametersForm(e.state, &e.selectedParametersIdx, onModify, onRebuild)
	case "Warmup":
		form = buildWarmupForm(e.state, &e.selectedWarmupsIdx, onModify, onRebuild)
	case "Response Cache":
		form = buildResponseCacheForm(e.state, onModify)
	case "Ensemble":
		form = buildEnsembleForm(e.state, onModify)
	case "Versions & Models":
		form = buildVersionsForm(e.window, e.state, onModify)
	case "Export Repository":
		form = buildExportRepositoryForm(e.window, e.state)
	default:
		form = widget.NewLabel("Select a configuration section from the sidebar.")
	}

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
	path := e.state.GetModelFolderPath()
	if path == "" {
		path = "No folder opened"
	}
	dirtyStr := ""
	if e.state.IsDirty() {
		dirtyStr = " *"
	}
	e.statusLabel.SetText(fmt.Sprintf("Folder: %s%s", path, dirtyStr))
	e.statusLabel.Refresh()

	// Undo and redo removed

	// Update Validation Status & Live Preview Text
	e.validateCurrentConfig()
	e.updatePreviewText()
}

func (e *EditorUI) onNew() {
	dialog.ShowConfirm("Discard Changes?", "Are you sure you want to reset? Unsaved changes will be lost.", func(discard bool) {
		if discard {
			s := state.NewAppState()
			e.state.SetConfig(s.GetConfig())
			e.state.SetModelFolderPath("")
			e.state.SetDirty(false)
			e.reloadActiveSection()
		}
	}, e.window)
}

func (e *EditorUI) onOpen() {
	fd := dialog.NewFolderOpen(func(listable fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if listable == nil {
			return
		}

		path := listable.Path()
		e.LoadFolder(path)
		fyne.CurrentApp().Preferences().SetString("recent_folder", path)
	}, e.window)

	fd.Resize(fyne.NewSize(800, 550))
	fd.Show()
}

func (e *EditorUI) LoadFolder(path string) {
	configPath := filepath.Join(path, "config.pbtxt")
	cfg, err := fileio.LoadConfig(configPath)
	if err == nil {
		e.state.SetConfig(cfg)
		e.state.SetDirty(false)
	} else {
		// Initialize blank config named after the folder
		e.state.SetConfig(&model.ModelConfig{
			Name: filepath.Base(path),
		})
		e.state.SetDirty(true)
	}
	e.state.SetModelFolderPath(path)
	e.reloadActiveSection()
}

func (e *EditorUI) onSave() {
	path := e.state.GetModelFolderPath()
	if path == "" {
		dialog.ShowError(fmt.Errorf("please open a model folder first"), e.window)
		return
	}

	configPath := filepath.Join(path, "config.pbtxt")
	err := fileio.SaveConfig(configPath, e.state.GetConfig())
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	e.state.SetDirty(false)
	dialog.ShowInformation("Configuration Saved", "config.pbtxt saved successfully inside the model folder.", e.window)
}



type previewTheme struct {
	fyne.Theme
}

func (p *previewTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if n == theme.ColorNameInputBackground {
		return color.Transparent
	}
	if n == theme.ColorNameForeground {
		// Use the primary theme color for syntax
		return p.Theme.Color(theme.ColorNamePrimary, v)
	}
	return p.Theme.Color(n, v)
}

type readOnlyEntry struct {
	widget.Entry
}

func newReadOnlyEntry() *readOnlyEntry {
	e := &readOnlyEntry{}
	e.MultiLine = true
	e.ExtendBaseWidget(e)
	return e
}

func (e *readOnlyEntry) TypedRune(r rune) {
	// Block keyboard text insertion
}

func (e *readOnlyEntry) TypedKey(key *fyne.KeyEvent) {
	// Allow standard selection/navigation keys, block editing keys.
	// Copy (Ctrl+C/Cmd+C) and Select All (Ctrl+A/Cmd+A) shortcuts are handled via TypedShortcut.
	switch key.Name {
	case fyne.KeyUp, fyne.KeyDown, fyne.KeyLeft, fyne.KeyRight, fyne.KeyHome, fyne.KeyEnd, fyne.KeyPageUp, fyne.KeyPageDown:
		e.Entry.TypedKey(key)
	}
}

func (e *readOnlyEntry) TypedShortcut(shortcut fyne.Shortcut) {
	// Allow standard Copy and Select All shortcuts, discard modifying shortcuts
	switch shortcut.(type) {
	case *fyne.ShortcutCopy, *fyne.ShortcutSelectAll:
		e.Entry.TypedShortcut(shortcut)
	}
}
