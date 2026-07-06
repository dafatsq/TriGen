package fileio

import (
	"os"
	"path/filepath"

	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/model"
	"triton-config-studio/internal/parser"
)

// LoadConfig reads and parses a Triton config.pbtxt file
func LoadConfig(path string) (*model.ModelConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parser.Parse(string(bytes))
}

// SaveConfig generates and writes a Triton config.pbtxt file
func SaveConfig(path string, cfg *model.ModelConfig) error {
	content := generator.Generate(cfg)

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file atomically (write to temp then rename)
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}

	return nil
}
