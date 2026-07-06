package exporter

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/model"
)

// CalculateTVI calculates the Triton Version Integer based on semantic version major.minor.patch
// Formula: 10000 * major + 100 * minor + patch
func CalculateTVI(semVer string) (int64, error) {
	semVer = strings.TrimSpace(semVer)
	parts := strings.Split(semVer, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("semantic version must be in format major.minor.patch (e.g. 1.2.5)")
	}
	major, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid patch version: %w", err)
	}
	if major < 0 || minor < 0 || patch < 0 {
		return 0, fmt.Errorf("version numbers cannot be negative")
	}
	return 10000*major + 100*minor + patch, nil
}

// ExportRepository ZIPs the configuration and model file into a Triton model repository structure.
func ExportRepository(zipPath string, cfg *model.ModelConfig, modelFilePath string, tvi int64) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	modelName := strings.TrimSpace(cfg.Name)
	if modelName == "" {
		modelName = "model"
	}

	// 1. Generate config.pbtxt
	configText := generator.Generate(cfg)
	configPathInZip := filepath.Join(modelName, "config.pbtxt")
	w1, err := archive.Create(configPathInZip)
	if err != nil {
		return fmt.Errorf("failed to create config.pbtxt in zip: %w", err)
	}
	if _, err := io.WriteString(w1, configText); err != nil {
		return fmt.Errorf("failed to write config.pbtxt: %w", err)
	}

	// 2. Add model file under TVI/
	if modelFilePath != "" {
		modelFile, err := os.Open(modelFilePath)
		if err != nil {
			return fmt.Errorf("failed to open model file: %w", err)
		}
		defer modelFile.Close()

		modelBaseName := filepath.Base(modelFilePath)
		modelPathInZip := filepath.Join(modelName, strconv.FormatInt(tvi, 10), modelBaseName)
		w2, err := archive.Create(modelPathInZip)
		if err != nil {
			return fmt.Errorf("failed to create model entry in zip: %w", err)
		}
		if _, err := io.Copy(w2, modelFile); err != nil {
			return fmt.Errorf("failed to write model file to zip: %w", err)
		}
	}

	return nil
}

// ZipDirectory recursively packages the entire directory structure into a ZIP file.
func ZipDirectory(dirPath, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	baseDir := filepath.Base(dirPath)

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		zipEntryPath := filepath.Join(baseDir, relPath)
		w, err := archive.Create(zipEntryPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return err
	})

	return err
}
