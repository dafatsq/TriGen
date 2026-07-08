package ui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestInitializedSelectDoesNotFireChangeHandler(t *testing.T) {
	test.NewTempApp(t)

	called := false
	selectWidget := newInitializedSelect([]string{"TYPE_FP32", "TYPE_INT32"}, "TYPE_FP32", func(string) {
		called = true
	})

	if called {
		t.Fatal("initial select value should not be treated as a user modification")
	}

	selectWidget.OnChanged("TYPE_INT32")
	if !called {
		t.Fatal("change handler should still run for later user modifications")
	}
}

func TestParsePositiveInt32ListRejectsInvalidTokens(t *testing.T) {
	if _, err := parsePositiveInt32List("4, typo, 8"); err == nil {
		t.Fatal("invalid preferred batch size tokens should be reported")
	}
	if _, err := parsePositiveInt32List("4, 0, 8"); err == nil {
		t.Fatal("non-positive preferred batch sizes should be rejected")
	}
}

func TestParsePositiveInt64ListRejectsInvalidTokens(t *testing.T) {
	if _, err := parsePositiveInt64List("1, typo, 5"); err == nil {
		t.Fatal("invalid version tokens should be reported")
	}
	if _, err := parsePositiveInt64List("1, -1, 5"); err == nil {
		t.Fatal("non-positive version numbers should be rejected")
	}
}

func TestCloseAndRemoveFileRemovesPartialExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create partial export: %v", err)
	}

	closeAndRemoveFile(file, path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("partial export should be removed, stat err: %v", err)
	}
}
