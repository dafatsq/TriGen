package exporter

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"triton-config-studio/internal/model"
)

func TestCalculateTVI(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1.2.5", 10205, false},
		{"0.0.1", 1, false},
		{"10.20.5", 102005, false},
		{"2.0.0", 20000, false},
		{"1.2", 0, true},
		{"1.2.3.4", 0, true},
		{"a.b.c", 0, true},
		{"-1.2.3", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := CalculateTVI(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("for input %q: expected %d, got %d", tt.input, tt.expected, got)
			}
		}
	}
}

func TestCalculateTVIRejectsAmbiguousVersions(t *testing.T) {
	for _, semver := range []string{"1.100.0", "1.2.100"} {
		if _, err := CalculateTVI(semver); err == nil {
			t.Fatalf("CalculateTVI(%q) should reject minor/patch values above 99", semver)
		}
	}
}

func TestZipEntryNameUsesSlashSeparators(t *testing.T) {
	got, err := zipEntryName("model", "10205\\model.onnx")
	if err != nil {
		t.Fatalf("zipEntryName returned error: %v", err)
	}
	if got != "model/10205/model.onnx" {
		t.Fatalf("expected slash-separated ZIP entry, got %q", got)
	}
}

func TestExportRepositoryRejectsUnsafeModelName(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "repo.zip")

	err := ExportRepository(zipPath, &model.ModelConfig{Name: "../evil"}, "", 1)
	if err == nil {
		t.Fatal("ExportRepository should reject model names that escape the ZIP root")
	}
}

func TestZipDirectoryRejectsOutputInsideSource(t *testing.T) {
	modelDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(modelDir, "config.pbtxt"), []byte("config"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	err := ZipDirectory(modelDir, filepath.Join(modelDir, "repo.zip"))
	if err == nil {
		t.Fatal("ZipDirectory should reject output paths inside the source directory")
	}
}

func TestExportRepository(t *testing.T) {
	// Create a dummy model file
	tmpDir, err := os.MkdirTemp("", "triton_export_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyModelPath := filepath.Join(tmpDir, "model.onnx")
	dummyModelContent := []byte("dummy model data")
	if err := os.WriteFile(dummyModelPath, dummyModelContent, 0644); err != nil {
		t.Fatalf("failed to write dummy model file: %v", err)
	}

	zipPath := filepath.Join(tmpDir, "repo.zip")
	cfg := &model.ModelConfig{
		Name:         "test_model",
		MaxBatchSize: 16,
	}
	tvi := int64(10205)

	err = ExportRepository(zipPath, cfg, dummyModelPath, tvi)
	if err != nil {
		t.Fatalf("ExportRepository failed: %v", err)
	}

	// Verify the zip file contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open generated zip: %v", err)
	}
	defer r.Close()

	expectedFiles := map[string]string{
		"test_model/config.pbtxt":     "name: \"test_model\"\nmax_batch_size: 16\n",
		"test_model/10205/model.onnx": "dummy model data",
	}

	for _, f := range r.File {
		expectedContent, exists := expectedFiles[f.Name]
		if !exists {
			t.Errorf("unexpected file in zip: %s", f.Name)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Errorf("failed to open file in zip %s: %v", f.Name, err)
			continue
		}
		contentBytes, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Errorf("failed to read file in zip %s: %v", f.Name, err)
			continue
		}

		if string(contentBytes) != expectedContent {
			t.Errorf("content mismatch for %s: expected %q, got %q", f.Name, expectedContent, string(contentBytes))
		}
		delete(expectedFiles, f.Name)
	}

	if len(expectedFiles) > 0 {
		t.Errorf("missing expected files in zip: %v", expectedFiles)
	}
}

func TestZipDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "zip_dir_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a dummy repository structure
	modelDir := filepath.Join(tmpDir, "mnist_model")
	if err := os.MkdirAll(filepath.Join(modelDir, "10205"), 0755); err != nil {
		t.Fatalf("failed to create version dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(modelDir, "config.pbtxt"), []byte("config content"), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	err = os.WriteFile(filepath.Join(modelDir, "10205", "model.onnx"), []byte("model binary content"), 0644)
	if err != nil {
		t.Fatalf("failed to write model file: %v", err)
	}

	zipPath := filepath.Join(tmpDir, "mnist_model.zip")
	err = ZipDirectory(modelDir, zipPath)
	if err != nil {
		t.Fatalf("ZipDirectory failed: %v", err)
	}

	// Verify the zip contents
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open generated zip: %v", err)
	}
	defer r.Close()

	expectedFiles := map[string]string{
		"mnist_model/config.pbtxt":     "config content",
		"mnist_model/10205/model.onnx": "model binary content",
	}

	for _, f := range r.File {
		expectedContent, exists := expectedFiles[f.Name]
		if !exists {
			t.Errorf("unexpected file in zip: %s", f.Name)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			t.Errorf("failed to open file %s in zip: %v", f.Name, err)
			continue
		}
		contentBytes, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Errorf("failed to read file %s in zip: %v", f.Name, err)
			continue
		}

		if string(contentBytes) != expectedContent {
			t.Errorf("content mismatch for %s: expected %q, got %q", f.Name, expectedContent, string(contentBytes))
		}
		delete(expectedFiles, f.Name)
	}

	if len(expectedFiles) > 0 {
		t.Errorf("missing expected files in zip: %v", expectedFiles)
	}
}
