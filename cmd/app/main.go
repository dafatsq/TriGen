package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"triton-config-studio/internal/state"
	"triton-config-studio/internal/ui"
)

func main() {
	// 1. Create Fyne Application
	a := app.New()

	// 2. Set custom premium Slate & Cyan theme
	a.Settings().SetTheme(&ui.StudioTheme{})

	// 3. Create window
	w := a.NewWindow("Triton Config Studio")

	// 4. Initialize state and UI
	s := state.NewAppState()
	editor := ui.NewEditorUI(w, s)

	w.SetContent(editor.Build())
	w.Resize(fyne.NewSize(1024, 768))

	// 5. Run desktop app
	w.ShowAndRun()
}
