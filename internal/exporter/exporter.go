package exporter

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"triton-config-studio/internal/generator"
	"triton-config-studio/internal/model"
)

// CalculateTVI calculates the Triton Version Integer based on semantic version major.minor.patch.
// Formula: 10000 * major + 100 * minor + patch. Minor and patch must be <= 99 to avoid collisions.
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
	if minor > 99 || patch > 99 {
		return 0, fmt.Errorf("minor and patch version numbers must be between 0 and 99 to avoid TVI collisions")
	}
	return 10000*major + 100*minor + patch, nil
}

// ExportRepository ZIPs the configuration and model file into a Triton model repository structure.
func ExportRepository(zipPath string, cfg *model.ModelConfig, modelFilePath string, tvi int64) (err error) {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func() {
		if closeErr := zipFile.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close zip file: %w", closeErr)
		}
	}()

	archive := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := archive.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to finalize zip archive: %w", closeErr)
		}
	}()

	modelName, err := safeZipSegment(cfg.Name, "model")
	if err != nil {
		return err
	}

	// 1. Generate config.pbtxt
	configText := generator.Generate(cfg)
	configPathInZip, err := zipEntryName(modelName, "config.pbtxt")
	if err != nil {
		return err
	}
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

		modelBaseName, err := safeZipSegment(filepath.Base(modelFilePath), "model.bin")
		if err != nil {
			_ = modelFile.Close()
			return err
		}
		modelPathInZip, err := zipEntryName(modelName, path.Join(strconv.FormatInt(tvi, 10), modelBaseName))
		if err != nil {
			_ = modelFile.Close()
			return err
		}
		w2, err := archive.Create(modelPathInZip)
		if err != nil {
			_ = modelFile.Close()
			return fmt.Errorf("failed to create model entry in zip: %w", err)
		}
		if _, err := io.Copy(w2, modelFile); err != nil {
			_ = modelFile.Close()
			return fmt.Errorf("failed to write model file to zip: %w", err)
		}
		if err := modelFile.Close(); err != nil {
			return fmt.Errorf("failed to close model file: %w", err)
		}
	}

	return nil
}

// ZipDirectory recursively packages the entire directory structure into a ZIP file.
func ZipDirectory(dirPath, zipPath string) (err error) {
	if err := ensureZipOutsideSource(dirPath, zipPath); err != nil {
		return err
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer func() {
		if closeErr := zipFile.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close zip file: %w", closeErr)
		}
	}()

	return ZipDirectoryToWriter(dirPath, zipFile)
}

// ZipDirectoryToWriter recursively packages dirPath into writer as a ZIP archive.
func ZipDirectoryToWriter(dirPath string, writer io.Writer) (err error) {
	archive := zip.NewWriter(writer)
	defer func() {
		if closeErr := archive.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("failed to finalize zip archive: %w", closeErr)
		}
	}()

	baseDir, err := safeZipSegment(filepath.Base(dirPath), "repository")
	if err != nil {
		return err
	}

	return filepath.Walk(dirPath, func(pathOnDisk string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(pathOnDisk)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, pathOnDisk)
		if err != nil {
			_ = file.Close()
			return err
		}
		zipEntryPath, err := zipEntryName(baseDir, relPath)
		if err != nil {
			_ = file.Close()
			return err
		}

		w, err := archive.Create(zipEntryPath)
		if err != nil {
			_ = file.Close()
			return err
		}

		if _, err := io.Copy(w, file); err != nil {
			_ = file.Close()
			return err
		}
		return file.Close()
	})
}

func ensureZipOutsideSource(dirPath, zipPath string) error {
	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	absZip, err := filepath.Abs(zipPath)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(absDir, absZip)
	if err != nil {
		return err
	}
	if rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..") {
		return fmt.Errorf("zip output path must be outside the source directory")
	}
	return nil
}

func safeZipSegment(value, fallback string) (string, error) {
	segment := strings.TrimSpace(value)
	if segment == "" {
		segment = fallback
	}
	if segment == "." || segment == ".." || strings.Contains(segment, "/") || strings.Contains(segment, "\\") {
		return "", fmt.Errorf("unsafe ZIP path segment %q", value)
	}
	return segment, nil
}

func zipEntryName(baseDir, relPath string) (string, error) {
	relPath = strings.ReplaceAll(filepath.ToSlash(relPath), "\\", "/")
	cleanRel := path.Clean(relPath)
	if cleanRel == "." || path.IsAbs(cleanRel) || cleanRel == ".." || strings.HasPrefix(cleanRel, "../") {
		return "", fmt.Errorf("unsafe ZIP entry path %q", relPath)
	}
	return path.Join(baseDir, cleanRel), nil
}
