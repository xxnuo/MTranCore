package main

import (
	"context"
	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/logger"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// EngineManager manages the translation engine lifecycle
type EngineManager struct {
	translator  *engine.Translator
	loadedFiles *engine.LoadedFiles
	queue       *TranslationQueue
	mu          sync.RWMutex
	config      *Config
	closeOnce   sync.Once
}

// LoadOptions configuration for loading the engine
type LoadOptions struct {
	Path                 string
	ModelPath            string
	LexicalShortlistPath string
	VocabularyPath       string
	VocabularyPaths      []string
}

// LoadResult represents the result of a load operation
type LoadResult struct {
	Success       bool
	ErrorCode     ErrorCode
	ErrorMessage  string
	AlreadyLoaded bool
}

// NewEngineManager creates a new engine manager instance
func NewEngineManager(cfg *Config) *EngineManager {
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	em.autoLoad()
	return em
}
func (em *EngineManager) autoLoad() {
	req := LoadOptions{}
	if em.config.ModelPath != "" {
		req.Path = em.config.ModelPath
	}
	if em.config.ModelFile != "" {
		req.ModelPath = em.config.ModelFile
	}
	if em.config.ShortlistFile != "" {
		req.LexicalShortlistPath = em.config.ShortlistFile
	}
	if em.config.VocabFile != "" {
		req.VocabularyPath = em.config.VocabFile
	}
	if len(em.config.VocabFiles) > 0 {
		req.VocabularyPaths = em.config.VocabFiles
	}
	if req.Path == "" && req.ModelPath == "" {
		logger.Fatal("No model configuration found. Set MODEL_PATH or MODEL_FILE environment variable")
	}
	ctx := context.Background()
	result := em.Load(ctx, req)
	if !result.Success {
		logger.Fatal("Auto-load failed: %s", result.ErrorMessage)
	}
	logger.Info("Model auto-loaded successfully")
}

// ResolvePath resolves a path (absolute or relative to work directory)
func (em *EngineManager) ResolvePath(path string) (string, ErrorCode, string) {
	if path == "" {
		return "", CodeLoadInvalidParams, "path is required"
	}
	var fullPath string
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		fullPath = filepath.Join(em.config.WorkDir, path)
	}
	// Check if path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", CodeLoadPathNotExists, "path does not exist: " + fullPath
	}
	return fullPath, CodeSuccess, ""
}

// resolveFilePath resolves a file path (absolute or relative to work directory)
func (em *EngineManager) resolveFilePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(em.config.WorkDir, path)
}

// Load loads the translation engine with the LoadOptions
func (em *EngineManager) Load(ctx context.Context, req LoadOptions) LoadResult {
	logger.Debug("[DEBUG-ENGINE] Load: starting, req=%+v", req)
	em.mu.Lock()
	defer em.mu.Unlock()
	// Check if engine is already loaded
	if em.translator != nil {
		logger.Debug("[DEBUG-ENGINE] Load: engine already loaded")
		return LoadResult{
			Success:       true,
			ErrorCode:     CodeSuccess,
			AlreadyLoaded: true,
		}
	}
	// Unload existing engine if any
	em.unloadEngineLocked()
	var config engine.EngineConfig
	// Priority: individual file paths > path
	if req.ModelPath != "" || req.LexicalShortlistPath != "" ||
		req.VocabularyPath != "" || len(req.VocabularyPaths) > 0 {
		// Use individual file paths
		if req.ModelPath == "" {
			return LoadResult{
				Success:      false,
				ErrorCode:    CodeLoadInvalidParams,
				ErrorMessage: "model_path is required when using individual file paths",
			}
		}
		if req.LexicalShortlistPath == "" {
			return LoadResult{
				Success:      false,
				ErrorCode:    CodeLoadInvalidParams,
				ErrorMessage: "lexical_shortlist_path is required when using individual file paths",
			}
		}
		// Merge vocabulary_path and vocabulary_paths
		vocabPaths := []string{}
		if req.VocabularyPath != "" {
			vocabPaths = append(vocabPaths, req.VocabularyPath)
		}
		vocabPaths = append(vocabPaths, req.VocabularyPaths...)
		if len(vocabPaths) == 0 {
			return LoadResult{
				Success:      false,
				ErrorCode:    CodeLoadInvalidParams,
				ErrorMessage: "at least one vocabulary path is required (vocabulary_path or vocabulary_paths)",
			}
		}
		// Resolve paths (make absolute if relative)
		modelPath := em.resolveFilePath(req.ModelPath)
		shortlistPath := em.resolveFilePath(req.LexicalShortlistPath)
		resolvedVocabPaths := make([]string, len(vocabPaths))
		for i, vp := range vocabPaths {
			resolvedVocabPaths[i] = em.resolveFilePath(vp)
		}
		config = engine.EngineConfig{
			ModelPath:            modelPath,
			LexicalShortlistPath: shortlistPath,
			VocabularyPaths:      resolvedVocabPaths,
			MaxLengthBreak:       em.config.MaxLengthBreak,
		}
	} else {
		// Use path (model directory)
		fullPath, errCode, errMsg := em.ResolvePath(req.Path)
		if errCode != CodeSuccess {
			return LoadResult{
				Success:      false,
				ErrorCode:    errCode,
				ErrorMessage: errMsg,
			}
		}
		config = engine.EngineConfig{
			ModelDir:       fullPath,
			MaxLengthBreak: em.config.MaxLengthBreak,
		}
	}
	// Create translator
	logger.Debug("[DEBUG-ENGINE] Load: calling engine.CreateTranslator")
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
		logger.Debug("[DEBUG-ENGINE] Load: CreateTranslator failed: %v", err)
		errMsg := err.Error()
		if containsAny(errMsg, "not found", "missing") {
			return LoadResult{
				Success:      false,
				ErrorCode:    CodeLoadIncompleteFiles,
				ErrorMessage: err.Error(),
			}
		}
		return LoadResult{
			Success:      false,
			ErrorCode:    CodeLoadInternalError,
			ErrorMessage: err.Error(),
		}
	}
	logger.Debug("[DEBUG-ENGINE] Load: CreateTranslator succeeded, translator=%v", translator)
	em.translator = translator
	em.loadedFiles = loadedFiles
	logger.Debug("[DEBUG-ENGINE] Load: em.translator set to %v", em.translator)
	// Update the queue's translator
	if em.queue != nil {
		logger.Debug("[DEBUG-ENGINE] Load: calling queue.SetTranslator")
		em.queue.SetTranslator(translator)
		logger.Debug("[DEBUG-ENGINE] Load: queue.SetTranslator done")
	} else {
		logger.Debug("[DEBUG-ENGINE] Load: queue is nil!")
	}
	logger.Debug("[DEBUG-ENGINE] Load: success")
	return LoadResult{
		Success:   true,
		ErrorCode: CodeSuccess,
	}
}

// WaitForIdle waits for active requests to complete or timeout
func (em *EngineManager) WaitForIdle(activeCounter *int32, timeoutDuration time.Duration) bool {
	timeout := time.After(timeoutDuration)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return false
		case <-ticker.C:
			if activeCounter == nil || atomic.LoadInt32(activeCounter) == 0 {
				return true
			}
		}
	}
}

// unloadEngineLocked closes the translator and loaded files (must hold lock)
// Note: This does NOT clear lastPoweronRequest, so it can be used for auto-reload
func (em *EngineManager) unloadEngineLocked() error {
	if em.translator != nil {
		if err := em.translator.Close(context.Background()); err != nil {
			logger.Warn("Failed to close translator: %v", err)
			return err
		}
		em.translator = nil
	}
	if em.loadedFiles != nil {
		em.loadedFiles.Close()
		em.loadedFiles = nil
	}
	// Clear the queue's translator
	if em.queue != nil {
		em.queue.SetTranslator(nil)
	}
	return nil
}

// Close closes the engine manager and releases resources
func (em *EngineManager) Close() error {
	var err error
	em.closeOnce.Do(func() {
		// Close the translation queue first
		if em.queue != nil {
			em.queue.Close()
		}
		em.mu.Lock()
		defer em.mu.Unlock()
		err = em.unloadEngineLocked()
	})
	return err
}

// GetTranslator returns the current translator (with read lock)
func (em *EngineManager) GetTranslator() *engine.Translator {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.translator
}

// IsReady checks if the engine is health
func (em *EngineManager) IsReady() bool {
	if em.queue != nil {
		return em.queue.IsReady()
	}
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.translator != nil
}

// GetQueue returns the translation queue
func (em *EngineManager) GetQueue() *TranslationQueue {
	logger.Debug("[DEBUG-ENGINE] GetQueue: returning queue=%v, queue.translator=%v", em.queue, func() interface{} {
		if em.queue == nil {
			return nil
		}
		em.queue.mu.RLock()
		defer em.queue.mu.RUnlock()
		return em.queue.translator
	}())
	return em.queue
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
