package main

import (
	"github.com/xxnuo/MTranCore/internal/logger"
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	engine "github.com/xxnuo/MTranCore/engine"
)

// EngineManager manages the translation engine lifecycle
type EngineManager struct {
	translator  *engine.Translator
	loadedFiles *engine.LoadedFiles
	queue       *TranslationQueue
	mu          sync.RWMutex
	config      *Config
	closeOnce   sync.Once
	// lastPoweronRequest stores the last successful poweron request for auto-reload after reboot
	lastPoweronRequest *PoweronRequest
}

// PoweronResult represents the result of a poweron operation
type PoweronResult struct {
	Success       bool
	ErrorCode     ErrorCode
	ErrorMessage  string
	AlreadyLoaded bool
}

// RebootResult represents the result of a reboot operation
type RebootResult struct {
	Success      bool
	ErrorCode    ErrorCode
	ErrorMessage string
}

// NewEngineManager creates a new engine manager instance
func NewEngineManager(cfg *Config) *EngineManager {
	return &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
}

// ResolvePath resolves a path (absolute or relative to work directory)
func (em *EngineManager) ResolvePath(path string) (string, ErrorCode, string) {
	if path == "" {
		return "", CodePoweronInvalidParams, "path is required"
	}

	var fullPath string
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		fullPath = filepath.Join(em.config.WorkDir, path)
	}

	// Check if path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", CodePoweronPathNotExists, "path does not exist: " + fullPath
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

// PoweronWithRequest loads the translation engine with the PoweronRequest
func (em *EngineManager) PoweronWithRequest(ctx context.Context, req PoweronRequest) PoweronResult {
	logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: starting, req=%+v", req)
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check if engine is already loaded
	if em.translator != nil {
		logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: engine already loaded")
		return PoweronResult{
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
			return PoweronResult{
				Success:      false,
				ErrorCode:    CodePoweronInvalidParams,
				ErrorMessage: "model_path is required when using individual file paths",
			}
		}
		if req.LexicalShortlistPath == "" {
			return PoweronResult{
				Success:      false,
				ErrorCode:    CodePoweronInvalidParams,
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
			return PoweronResult{
				Success:      false,
				ErrorCode:    CodePoweronInvalidParams,
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
						return PoweronResult{
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
	logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: calling engine.CreateTranslator")
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
		logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: CreateTranslator failed: %v", err)
		errMsg := err.Error()
		if containsAny(errMsg, "not found", "missing") {
			return PoweronResult{
				Success:      false,
				ErrorCode:    CodePoweronIncompleteFiles,
				ErrorMessage: err.Error(),
			}
		}
		return PoweronResult{
			Success:      false,
			ErrorCode:    CodePoweronInternalError,
			ErrorMessage: err.Error(),
		}
	}
	logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: CreateTranslator succeeded, translator=%v", translator)

	em.translator = translator
	em.loadedFiles = loadedFiles
	logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: em.translator set to %v", em.translator)

	// Save the poweron request for potential auto-reload after reboot
	em.lastPoweronRequest = &req

	// Update the queue's translator
	if em.queue != nil {
		logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: calling queue.SetTranslator")
		em.queue.SetTranslator(translator)
		logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: queue.SetTranslator done")
	} else {
		logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: queue is nil!")
	}

	logger.Debug("[DEBUG-ENGINE] PoweronWithRequest: success")
	return PoweronResult{
		Success:   true,
		ErrorCode: CodeSuccess,
	}
}

// Poweron loads the translation engine with model files (backward compatibility)
func (em *EngineManager) Poweron(ctx context.Context, path string) PoweronResult {
	return em.PoweronWithRequest(ctx, PoweronRequest{Path: path})
}

// Reboot reloads the translation engine
func (em *EngineManager) Reboot(ctx context.Context, force bool, activeCounter *int32) RebootResult {
	// If not force, wait for active requests to complete
	if !force {
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return RebootResult{
					Success:      false,
					ErrorCode:    CodeRebootWaitingTask,
					ErrorMessage: "Timeout waiting for active requests to complete",
				}
			case <-ticker.C:
				if activeCounter == nil || atomic.LoadInt32(activeCounter) == 0 {
					goto reboot
				}
			}
		}
	}

reboot:
	em.mu.Lock()
	defer em.mu.Unlock()

	// Save the last poweron request before unloading
	lastReq := em.lastPoweronRequest

	// Close existing translator and loaded files
	if err := em.unloadEngineLocked(); err != nil {
		return RebootResult{
			Success:      false,
			ErrorCode:    CodeRebootInternalError,
			ErrorMessage: "Failed to close translator: " + err.Error(),
		}
	}

	// Auto-reload engine if we have a saved poweron request
	if lastReq != nil {
		logger.Info("Auto-reloading engine after reboot...")
		result := em.PoweronWithRequest(ctx, *lastReq)
		if !result.Success {
			return RebootResult{
				Success:      false,
				ErrorCode:    CodeRebootInternalError,
				ErrorMessage: "Failed to reload engine after reboot: " + result.ErrorMessage,
			}
		}
		logger.Info("Engine auto-reloaded successfully after reboot")
	}

	return RebootResult{
		Success:   true,
		ErrorCode: CodeSuccess,
	}
}

// RebootAsync performs a reboot asynchronously with optional delay
func (em *EngineManager) RebootAsync(waitSeconds int, force bool, activeCounter *int32, onComplete func(RebootResult)) {
	go func() {
		// Wait for specified time
		if waitSeconds > 0 {
			time.Sleep(time.Duration(waitSeconds) * time.Second)
		}

		// If not force, wait for active requests to complete
		if !force {
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					logger.Warn("Reboot timeout reached, forcing reboot")
					goto reboot
				case <-ticker.C:
					if activeCounter == nil || atomic.LoadInt32(activeCounter) == 0 {
						goto reboot
					}
				}
			}
		}

	reboot:
		em.mu.Lock()
		
		// Save the last poweron request before unloading
		lastReq := em.lastPoweronRequest

		// Close existing translator and loaded files
		if err := em.unloadEngineLocked(); err != nil {
			logger.Error("Failed to close translator during reboot: %v", err)
			em.mu.Unlock()
			if onComplete != nil {
				onComplete(RebootResult{
					Success:      false,
					ErrorCode:    CodeRebootInternalError,
					ErrorMessage: "Failed to close translator: " + err.Error(),
				})
			}
			return
		}

		em.mu.Unlock()

		// Auto-reload engine if we have a saved poweron request
		if lastReq != nil {
			logger.Info("Auto-reloading engine after reboot...")
			ctx := context.Background()
			result := em.PoweronWithRequest(ctx, *lastReq)
			if !result.Success {
				logger.Error("Failed to reload engine after reboot: %s", result.ErrorMessage)
				if onComplete != nil {
					onComplete(RebootResult{
						Success:      false,
						ErrorCode:    CodeRebootInternalError,
						ErrorMessage: "Failed to reload engine after reboot: " + result.ErrorMessage,
					})
				}
				return
			}
			logger.Info("Engine auto-reloaded successfully after reboot")
		}

		if onComplete != nil {
			onComplete(RebootResult{
				Success:   true,
				ErrorCode: CodeSuccess,
			})
		}
	}()
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

// IsReady checks if the engine is ready
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
