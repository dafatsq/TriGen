package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"triton-config-studio/internal/state"
	"triton-config-studio/internal/validator"
)

func collectValidationIssues(s *state.AppState) []string {
	issues := validator.Validate(s.GetConfig())
	issues = append(issues, s.GetUIErrors()...)
	return issues
}

func blockingValidationIssues(s *state.AppState) []string {
	var blocking []string
	for _, issue := range collectValidationIssues(s) {
		if strings.HasPrefix(issue, "Error:") {
			blocking = append(blocking, issue)
		}
	}
	return blocking
}

func showValidationIssuesDialog(title string, issues []string, win fyne.Window) {
	var sb strings.Builder
	for _, issue := range issues {
		sb.WriteString("- " + issue + "\n")
	}
	scrollContent := container.NewVScroll(widget.NewLabel(sb.String()))
	scrollContent.SetMinSize(fyne.NewSize(500, 300))
	dialog.NewCustom(title, "Close", scrollContent, win).Show()
}
