package fileio

import (
	"os"
	"path/filepath"
	"testing"

	"triton-config-studio/internal/model"
)

func TestSaveConfigDoesNotUseFixedTempSibling(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.pbtxt")
	fixedTempSibling := target + ".tmp"

	if err := os.WriteFile(target, []byte("old config"), 0644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(fixedTempSibling, []byte("do not touch"), 0644); err != nil {
		t.Fatalf("write fixed temp sibling: %v", err)
	}

	err := SaveConfig(target, &model.ModelConfig{Name: "saved_model", MaxBatchSize: 4})
	if err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	want := "name: \"saved_model\"\nmax_batch_size: 4\n"
	if string(content) != want {
		t.Fatalf("target content mismatch:\nwant %q\n got %q", want, string(content))
	}

	tmpContent, err := os.ReadFile(fixedTempSibling)
	if err != nil {
		t.Fatalf("fixed temp sibling should remain untouched: %v", err)
	}
	if string(tmpContent) != "do not touch" {
		t.Fatalf("fixed temp sibling content changed: %q", string(tmpContent))
	}
}
