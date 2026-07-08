package fileio

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/model"
	"triton-config-studio/internal/parser"
)

// LoadConfig reads and parses a Triton config.pbtxt file.
func LoadConfig(path string) (*model.ModelConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadConfigFromReader(file)
}

// LoadConfigFromReader reads and parses a Triton config.pbtxt stream.
func LoadConfigFromReader(reader io.Reader) (*model.ModelConfig, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return parser.Parse(string(bytes))
}

// SaveConfig generates and writes a Triton config.pbtxt file.
func SaveConfig(path string, cfg *model.ModelConfig) error {
	content := generator.Generate(cfg)

	// Ensure the parent directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := replaceFile(tmpPath, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

// SaveConfigToWriter generates and writes a Triton config.pbtxt stream.
func SaveConfigToWriter(writer io.Writer, cfg *model.ModelConfig) error {
	_, err := io.WriteString(writer, generator.Generate(cfg))
	return err
}

func replaceFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if runtime.GOOS != "windows" {
		return err
	}

	// Windows cannot reliably rename over an existing file. Remove and retry there.
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(src, dst)
}
