package state

import (
	"encoding/json"
	"strings"
	"sync"

	"triton-config-studio/internal/model"
)

type AppState struct {
	mu           sync.Mutex
	config       *model.ModelConfig
	filePath     string
	isDirty      bool
	undoStack    []string // JSON strings
	redoStack    []string // JSON strings
	listeners    []func()
	uiErrors     map[string]string // UI input parsing errors
}

func NewAppState() *AppState {
	return &AppState{
		config: &model.ModelConfig{
			Name:         "new_model",
			MaxBatchSize: 0,
		},
		undoStack: []string{},
		redoStack: []string{},
		uiErrors:  make(map[string]string),
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

func (s *AppState) GetFilePath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.filePath
}

func (s *AppState) SetFilePath(path string) {
	s.mu.Lock()
	s.filePath = path
	s.mu.Unlock()
	s.notifyListeners()
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

// SaveSnapshot saves the current state of ModelConfig to the undo stack
func (s *AppState) SaveSnapshot() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(s.config)
	if err != nil {
		return
	}

	// Push current state to undo stack
	s.undoStack = append(s.undoStack, string(data))
	// Clear redo stack on new actions
	s.redoStack = nil
	s.isDirty = true
}

func (s *AppState) CanUndo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.undoStack) > 0
}

func (s *AppState) CanRedo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.redoStack) > 0
}

func (s *AppState) Undo() {
	s.mu.Lock()
	if len(s.undoStack) == 0 {
		s.mu.Unlock()
		return
	}

	// Marshal current state to push to redo
	currentData, err := json.Marshal(s.config)
	if err == nil {
		s.redoStack = append(s.redoStack, string(currentData))
	}

	// Pop from undo stack
	lastIdx := len(s.undoStack) - 1
	prevJSON := s.undoStack[lastIdx]
	s.undoStack = s.undoStack[:lastIdx]

	var prevCfg model.ModelConfig
	if err := json.Unmarshal([]byte(prevJSON), &prevCfg); err == nil {
		s.config = &prevCfg
		s.isDirty = true
	}
	s.mu.Unlock()
	s.notifyListeners()
}

func (s *AppState) Redo() {
	s.mu.Lock()
	if len(s.redoStack) == 0 {
		s.mu.Unlock()
		return
	}

	// Marshal current state to push to undo
	currentData, err := json.Marshal(s.config)
	if err == nil {
		s.undoStack = append(s.undoStack, string(currentData))
	}

	// Pop from redo stack
	lastIdx := len(s.redoStack) - 1
	nextJSON := s.redoStack[lastIdx]
	s.redoStack = s.redoStack[:lastIdx]

	var nextCfg model.ModelConfig
	if err := json.Unmarshal([]byte(nextJSON), &nextCfg); err == nil {
		s.config = &nextCfg
		s.isDirty = true
	}
	s.mu.Unlock()
	s.notifyListeners()
}
