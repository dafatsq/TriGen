package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
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
	sidebarContainer *fyne.Container
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
		window:                    win,
		state:                     s,
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
	// 1. Sidebar Container initialization with tight spacing layout
	e.sidebarContainer = container.New(&tightVBoxLayout{spacing: 1})

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
	openFileBtn := widget.NewButtonWithIcon("Open File (.pbtxt)", theme.FileIcon(), e.onOpenFile)
	openFolderBtn := widget.NewButtonWithIcon("Open Folder (Model Repository)", theme.FolderOpenIcon(), e.onOpenFolder)
	saveBtn := widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), e.onSave)

	// Togglable Preview button
	e.previewButton = widget.NewButtonWithIcon("Preview", theme.VisibilityIcon(), e.togglePreview)

	// Templates selection dropdown
	var templateNames []string
	templateNames = append(templateNames, "Load Template...")
	for _, t := range templates.BuiltInTemplates {
		templateNames = append(templateNames, t.Name)
	}
	var templateSelect *widget.Select
	templateSelect = widget.NewSelect(templateNames, func(selected string) {
		if selected == "Load Template..." || selected == "" {
			return
		}
		e.confirmDiscardIfDirty(func() {
			for _, t := range templates.BuiltInTemplates {
				if t.Name == selected {
					e.state.SetConfig(model.CloneConfig(t.Config))
					e.state.SetModelFolderPath("")
					e.state.SetDirty(true)
					e.reloadActiveSection()
					dialog.ShowInformation("Template Loaded", fmt.Sprintf("Loaded %s template configuration.", t.Name), e.window)
					templateSelect.SetSelected("Load Template...")
					break
				}
			}
		})
	})
	templateSelect.SetSelected("Load Template...")

	toolbar := container.NewHBox(
		newBtn, openFileBtn, openFolderBtn, saveBtn,
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
	e.statusLabel = widget.NewLabel("Workspace: Unsaved")
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
	sidebarScroll := container.NewVScroll(e.sidebarContainer)
	sidebarScroll.SetMinSize(fyne.NewSize(200, 0))
	e.mainSplit = container.NewHSplit(sidebarScroll, e.rightContainer)
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
	e.valErrors = collectValidationIssues(e.state)
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

	showValidationIssuesDialog("Validation Errors", e.valErrors, e.window)
}

func (e *EditorUI) updateSidebarSections() {
	if e.sidebarContainer == nil {
		return
	}
	e.sidebarContainer.Objects = nil

	// Add Configuration Header Card (lighter color, non-hoverable)
	e.sidebarContainer.Add(newSidebarHeaderWidget("CONFIGURATION"))

	// Add config items
	configItems := []string{
		"General", "Inputs", "Outputs", "Version Policy", "Instance Groups",
		"Dynamic Batching", "Sequence Batching", "Optimization", "Parameters",
		"Warmup", "Response Cache", "Ensemble",
	}

	// Verify active section still exists in our available set
	hasActive := false
	for _, item := range configItems {
		if item == e.activeSection {
			hasActive = true
			break
		}
	}

	for _, item := range configItems {
		name := item
		e.sidebarContainer.Add(newSidebarItemWidget(name, e.activeSection == name, func() {
			e.activeSection = name
			e.updateSidebarSections()
			e.reloadActiveSection()
		}))
	}

	// Add Repository sections only in Folder Mode
	if e.state.GetModelFolderPath() != "" {
		e.sidebarContainer.Add(newSidebarHeaderWidget("MODEL REPOSITORY"))
		repoItems := []string{"Versions & Models", "Export Repository"}
		for _, item := range repoItems {
			if item == e.activeSection {
				hasActive = true
			}
		}
		for _, item := range repoItems {
			name := item
			e.sidebarContainer.Add(newSidebarItemWidget(name, e.activeSection == name, func() {
				e.activeSection = name
				e.updateSidebarSections()
				e.reloadActiveSection()
			}))
		}
	}

	if !hasActive {
		e.activeSection = "General"
		e.reloadActiveSection()
		// Redraw to reflect the change to "General" selection
		e.updateSidebarSections()
		return
	}

	e.sidebarContainer.Refresh()
}

func (e *EditorUI) onStateChanged() {
	// Dynamically show/hide repository tabs depending on folder mode
	e.updateSidebarSections()

	// Update File Path and Dirty State
	path := ""
	prefix := "Workspace: Unsaved"
	if filePath := e.state.GetConfigFilePath(); filePath != "" {
		path = filePath
		prefix = "File"
	} else if folderPath := e.state.GetModelFolderPath(); folderPath != "" {
		path = folderPath
		prefix = "Folder"
	}

	dirtyStr := ""
	if e.state.IsDirty() {
		dirtyStr = " *"
	}
	if path != "" {
		e.statusLabel.SetText(fmt.Sprintf("%s: %s%s", prefix, path, dirtyStr))
	} else {
		e.statusLabel.SetText(fmt.Sprintf("%s%s", prefix, dirtyStr))
	}
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
			e.state.SetConfigFilePath("")
			e.state.SetDirty(false)
			e.reloadActiveSection()
		}
	}, e.window)
}

func (e *EditorUI) onOpenFile() {
	e.confirmDiscardIfDirty(e.showOpenFileDialog)
}

func (e *EditorUI) showOpenFileDialog() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if reader == nil {
			return
		}
		path := reader.URI().Path()
		cfg, err := fileio.LoadConfigFromReader(reader)
		if closeErr := reader.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			dialog.ShowError(fmt.Errorf("error loading pbtxt file: %w", err), e.window)
			return
		}

		e.state.SetConfig(cfg)
		e.state.SetConfigFilePath(path)
		e.state.SetDirty(false)
		e.reloadActiveSection()
		fyne.CurrentApp().Preferences().SetString("recent_mode", "file")
		fyne.CurrentApp().Preferences().SetString("recent_path", path)
	}, e.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pbtxt"}))
	fd.Resize(fyne.NewSize(800, 550))
	fd.Show()
}

func (e *EditorUI) onOpenFolder() {
	e.confirmDiscardIfDirty(e.showOpenFolderDialog)
}

func (e *EditorUI) showOpenFolderDialog() {
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
		fyne.CurrentApp().Preferences().SetString("recent_mode", "folder")
		fyne.CurrentApp().Preferences().SetString("recent_path", path)
	}, e.window)

	fd.Resize(fyne.NewSize(800, 550))
	fd.Show()
}

func (e *EditorUI) confirmDiscardIfDirty(action func()) {
	if !e.state.IsDirty() {
		action()
		return
	}
	dialog.ShowConfirm("Discard Changes?", "Unsaved changes will be lost. Continue?", func(discard bool) {
		if discard {
			action()
		}
	}, e.window)
}

func (e *EditorUI) LoadFile(path string) {
	cfg, err := fileio.LoadConfig(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("error loading pbtxt file: %w", err), e.window)
		return
	}
	e.state.SetConfig(cfg)
	e.state.SetConfigFilePath(path)
	e.state.SetDirty(false)
	e.reloadActiveSection()
}

func (e *EditorUI) LoadFolder(path string) {
	configPath := filepath.Join(path, "config.pbtxt")
	cfg, err := fileio.LoadConfig(configPath)
	if err == nil {
		e.state.SetConfig(cfg)
		e.state.SetDirty(false)
	} else if os.IsNotExist(err) {
		// Initialize blank config named after the folder when no config exists yet.
		e.state.SetConfig(&model.ModelConfig{
			Name: filepath.Base(path),
		})
		e.state.SetDirty(true)
	} else {
		dialog.ShowError(fmt.Errorf("error loading config.pbtxt: %w", err), e.window)
		return
	}
	e.state.SetModelFolderPath(path)
	e.reloadActiveSection()
}

func (e *EditorUI) onSave() {
	e.validateCurrentConfig()
	if validator.HasBlockingErrors(e.valErrors) {
		showValidationIssuesDialog("Fix Validation Errors Before Saving", e.valErrors, e.window)
		return
	}

	configPath := e.state.GetConfigFilePath()
	modelFolderPath := e.state.GetModelFolderPath()

	if configPath != "" {
		err := fileio.SaveConfig(configPath, e.state.GetConfig())
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		e.state.SetDirty(false)
		dialog.ShowInformation("Configuration Saved", "config.pbtxt saved successfully.", e.window)
		return
	}

	if modelFolderPath != "" {
		targetPath := filepath.Join(modelFolderPath, "config.pbtxt")
		err := fileio.SaveConfig(targetPath, e.state.GetConfig())
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		e.state.SetDirty(false)
		dialog.ShowInformation("Configuration Saved", "config.pbtxt saved successfully inside the model folder.", e.window)
		return
	}

	// Unsaved new config, show Save File dialog
	fd := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if writer == nil {
			return
		}
		path := writer.URI().Path()
		err = fileio.SaveConfigToWriter(writer, e.state.GetConfig())
		if closeErr := writer.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}

		e.state.SetConfigFilePath(path)
		e.state.SetDirty(false)
		fyne.CurrentApp().Preferences().SetString("recent_mode", "file")
		fyne.CurrentApp().Preferences().SetString("recent_path", path)
		dialog.ShowInformation("Configuration Saved", "config.pbtxt saved successfully.", e.window)
	}, e.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".pbtxt"}))
	fd.SetFileName("config.pbtxt")
	fd.Resize(fyne.NewSize(800, 550))
	fd.Show()
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

// Custom widgets for the sidebar layout to support customized hover and non-hoverable elements

type sidebarHeaderWidget struct {
	widget.BaseWidget
	text  string
	label *canvas.Text
	bg    *canvas.Rectangle
}

func newSidebarHeaderWidget(text string) *sidebarHeaderWidget {
	label := canvas.NewText(text, theme.ForegroundColor())
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.TextSize = theme.TextSize()
	w := &sidebarHeaderWidget{
		text:  text,
		bg:    canvas.NewRectangle(color.RGBA{R: 28, G: 37, B: 53, A: 255}),
		label: label,
	}
	w.ExtendBaseWidget(w)
	return w
}

func customPad(object fyne.CanvasObject, top, bottom, left, right float32) fyne.CanvasObject {
	var topSpacer, bottomSpacer, leftSpacer, rightSpacer fyne.CanvasObject
	if top > 0 {
		t := canvas.NewRectangle(color.Transparent)
		t.SetMinSize(fyne.NewSize(0, top))
		topSpacer = t
	}
	if bottom > 0 {
		b := canvas.NewRectangle(color.Transparent)
		b.SetMinSize(fyne.NewSize(0, bottom))
		bottomSpacer = b
	}
	if left > 0 {
		l := canvas.NewRectangle(color.Transparent)
		l.SetMinSize(fyne.NewSize(left, 0))
		leftSpacer = l
	}
	if right > 0 {
		r := canvas.NewRectangle(color.Transparent)
		r.SetMinSize(fyne.NewSize(right, 0))
		rightSpacer = r
	}
	return container.NewBorder(topSpacer, bottomSpacer, leftSpacer, rightSpacer, object)
}

func (w *sidebarHeaderWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(
		w.bg,
		customPad(w.label, 5, 5, 8, 0),
	))
}

func (w *sidebarHeaderWidget) MinSize() fyne.Size {
	return customPad(w.label, 5, 5, 8, 0).MinSize()
}

type sidebarItemWidget struct {
	widget.BaseWidget
	text     string
	onTap    func()
	bg       *canvas.Rectangle
	label    *canvas.Text
	selected bool
}

func newSidebarItemWidget(text string, selected bool, onTap func()) *sidebarItemWidget {
	label := canvas.NewText("  "+text, theme.ForegroundColor())
	label.TextSize = theme.TextSize()
	w := &sidebarItemWidget{
		text:     text,
		onTap:    onTap,
		selected: selected,
		bg:       canvas.NewRectangle(color.Transparent),
		label:    label,
	}
	w.ExtendBaseWidget(w)
	w.updateStyle()
	return w
}

func (w *sidebarItemWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(
		w.bg,
		customPad(w.label, 3, 3, 6, 0),
	))
}

func (w *sidebarItemWidget) MinSize() fyne.Size {
	return customPad(w.label, 3, 3, 6, 0).MinSize()
}

func (w *sidebarItemWidget) Tapped(e *fyne.PointEvent) {
	if w.onTap != nil {
		w.onTap()
	}
}

func (w *sidebarItemWidget) MouseIn(e *desktop.MouseEvent) {
	if !w.selected {
		w.bg.FillColor = theme.HoverColor()
		w.bg.Refresh()
	}
}

func (w *sidebarItemWidget) MouseOut() {
	if !w.selected {
		w.bg.FillColor = color.Transparent
		w.bg.Refresh()
	}
}

func (w *sidebarItemWidget) updateStyle() {
	w.label.Color = theme.ForegroundColor()
	w.label.TextSize = theme.TextSize()
	if w.selected {
		w.bg.FillColor = theme.SelectionColor()
		w.label.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		w.bg.FillColor = color.Transparent
		w.label.TextStyle = fyne.TextStyle{}
	}
	w.bg.Refresh()
	w.label.Refresh()
}

// tightVBoxLayout lays out items vertically with pixel-perfect control over inter-item spacing
type tightVBoxLayout struct {
	spacing float32
}

func (d *tightVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	posY := float32(0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		childHeight := child.MinSize().Height
		child.Move(fyne.NewPos(0, posY))
		child.Resize(fyne.NewSize(size.Width, childHeight))
		posY += childHeight + d.spacing
	}
}

func (d *tightVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	width := float32(0)
	height := float32(0)
	visibleCount := 0
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		min := child.MinSize()
		if min.Width > width {
			width = min.Width
		}
		height += min.Height
		visibleCount++
	}
	if visibleCount > 1 {
		height += float32(visibleCount-1) * d.spacing
	}
	return fyne.NewSize(width, height)
}
