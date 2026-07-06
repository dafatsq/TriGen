package state

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"triton-config-studio/internal/model"
)

type ModelVersion struct {
	Version int64
	File    string
}

type AppState struct {
	mu              sync.Mutex
	config          *model.ModelConfig
	modelFolderPath string
	configFilePath   string
	isDirty         bool
	listeners       []func()
	uiErrors        map[string]string // UI input parsing errors
	semVersion      string            // Semantic version (e.g. 1.2.5)
	modelPath       string            // Uploaded model path
	versions        []ModelVersion
}

func NewAppState() *AppState {
	return &AppState{
		config: &model.ModelConfig{
			Name:         "new_model",
			MaxBatchSize: 0,
		},
		uiErrors:   make(map[string]string),
		semVersion: "1.0.0",
	}
}

func (s *AppState) SetUIError(key, msg string) {
	s.mu.Lock()
	if s.uiErrors == nil {
		s.uiErrors = make(map[string]string)
	}
	s.uiErrors[key] = msg
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) ClearUIError(key string) {
	s.mu.Lock()
	if s.uiErrors != nil {
		delete(s.uiErrors, key)
	}
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) ClearUIErrorsWithPrefix(prefix string) {
	s.mu.Lock()
	if s.uiErrors != nil {
		for k := range s.uiErrors {
			if strings.HasPrefix(k, prefix) {
				delete(s.uiErrors, k)
			}
		}
	}
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) GetUIErrors() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []string
	for _, v := range s.uiErrors {
		errs = append(errs, v)
	}
	return errs
}

func (s *AppState) GetConfig() *model.ModelConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config
}

func (s *AppState) SetConfig(cfg *model.ModelConfig) {
	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) GetModelFolderPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modelFolderPath
}

func (s *AppState) SetModelFolderPath(path string) {
	s.mu.Lock()
	s.modelFolderPath = path
	s.configFilePath = ""
	s.mu.Unlock()
	s.ScanVersions()
	s.notifyListeners()
}

func (s *AppState) GetConfigFilePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.configFilePath
}

func (s *AppState) SetConfigFilePath(path string) {
	s.mu.Lock()
	s.configFilePath = path
	s.modelFolderPath = ""
	s.versions = nil
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) GetVersions() []ModelVersion {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.versions
}

func (s *AppState) ScanVersions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.modelFolderPath == "" {
		s.versions = nil
		return
	}

	entries, err := os.ReadDir(s.modelFolderPath)
	if err != nil {
		s.versions = nil
		return
	}

	var detected []ModelVersion
	for _, entry := range entries {
		if entry.IsDir() {
			val, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && val >= 0 {
				// Scan this version folder for a model file
				subPath := filepath.Join(s.modelFolderPath, entry.Name())
				subEntries, err := os.ReadDir(subPath)
				fileName := "No binary found"
				if err == nil {
					for _, se := range subEntries {
						if !se.IsDir() {
							fileName = se.Name()
							break
						}
					}
				}
				detected = append(detected, ModelVersion{
					Version: val,
					File:    fileName,
				})
			}
		}
	}
	s.versions = detected
}

func (s *AppState) IsDirty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isDirty
}

func (s *AppState) SetDirty(dirty bool) {
	s.mu.Lock()
	s.isDirty = dirty
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) RegisterListener(f func()) {
	s.mu.Lock()
	s.listeners = append(s.listeners, f)
	s.mu.Unlock()
}

func (s *AppState) notifyListeners() {
	for _, f := range s.listeners {
		f()
	}
}

func (s *AppState) GetSemVersion() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.semVersion
}

func (s *AppState) SetSemVersion(v string) {
	s.mu.Lock()
	s.semVersion = v
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) GetModelPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.modelPath
}

func (s *AppState) SetModelPath(path string) {
	s.mu.Lock()
	s.modelPath = path
	s.mu.Unlock()
	s.notifyListeners()
}


