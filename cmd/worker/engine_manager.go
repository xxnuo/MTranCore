package main

import (
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

// Poweron loads the translation engine with model files
func (em *EngineManager) Poweron(ctx context.Context, path string) PoweronResult {
	fullPath, errCode, errMsg := em.ResolvePath(path)
	if errCode != CodeSuccess {
		return PoweronResult{
			Success:      false,
			ErrorCode:    errCode,
			ErrorMessage: errMsg,
		}
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	// Check if engine is already loaded
	if em.translator != nil {
		return PoweronResult{
			Success:       true,
			ErrorCode:     CodeSuccess,
			AlreadyLoaded: true,
		}
	}

	// Unload existing engine if any
	em.unloadEngineLocked()

	// Create translator using model directory
	config := engine.EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
		// Determine error code based on error message
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

	em.translator = translator
	em.loadedFiles = loadedFiles

	// Update the queue's translator
	if em.queue != nil {
		em.queue.SetTranslator(translator)
	}

	return PoweronResult{
		Success:   true,
		ErrorCode: CodeSuccess,
	}
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

	// Close existing translator and loaded files
	if err := em.unloadEngineLocked(); err != nil {
		return RebootResult{
			Success:      false,
			ErrorCode:    CodeRebootInternalError,
			ErrorMessage: "Failed to close translator: " + err.Error(),
		}
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
					Warn("Reboot timeout reached, forcing reboot")
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
		defer em.mu.Unlock()

		// Close existing translator and loaded files
		if err := em.unloadEngineLocked(); err != nil {
			Error("Failed to close translator during reboot: %v", err)
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
func (em *EngineManager) unloadEngineLocked() error {
	if em.translator != nil {
		if err := em.translator.Close(context.Background()); err != nil {
			Warn("Failed to close translator: %v", err)
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
