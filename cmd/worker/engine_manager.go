package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/wasm"
)

// EngineConfig represents the simplified configuration for loading an engine
type EngineConfig struct {
	// ModelDir is the directory containing all model files
	// If set, individual file paths are ignored and auto-discovered
	ModelDir string `json:"model_dir,omitempty"`

	// Individual file paths (used when ModelDir is not set)
	ModelPath            string   `json:"model_path,omitempty"`
	LexicalShortlistPath string   `json:"lexical_shortlist_path,omitempty"`
	VocabularyPaths      []string `json:"vocabulary_paths,omitempty"`
}

// LoadedFiles tracks opened files for cleanup
type LoadedFiles struct {
	ModelFile     *os.File
	ShortlistFile *os.File
	VocabFiles    []*os.File
}

// Close closes all opened files
func (lf *LoadedFiles) Close() {
	if lf.ModelFile != nil {
		lf.ModelFile.Close()
	}
	if lf.ShortlistFile != nil {
		lf.ShortlistFile.Close()
	}
	for _, f := range lf.VocabFiles {
		if f != nil {
			f.Close()
		}
	}
}

// CreateTranslator creates a new translator from the engine configuration
// This is the shared logic used by both HTTP and gRPC servers
func CreateTranslator(ctx context.Context, cfg EngineConfig) (*engine.Translator, *LoadedFiles, error) {
	var modelPath, shortlistPath string
	var vocabPaths []string
	var err error

	// Auto-discover model files if ModelDir is provided
	if cfg.ModelDir != "" {
		modelPath, shortlistPath, vocabPaths, err = discoverModelFiles(cfg.ModelDir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to discover model files: %w", err)
		}
	} else {
		// Use individual paths
		modelPath = cfg.ModelPath
		shortlistPath = cfg.LexicalShortlistPath
		vocabPaths = cfg.VocabularyPaths

		// Validate required fields
		if modelPath == "" {
			return nil, nil, fmt.Errorf("model_path is required")
		}
		if shortlistPath == "" {
			return nil, nil, fmt.Errorf("lexical_shortlist_path is required")
		}
		if len(vocabPaths) == 0 {
			return nil, nil, fmt.Errorf("at least one vocabulary_path is required")
		}
	}

	loadedFiles := &LoadedFiles{}

	// Open model file
	loadedFiles.ModelFile, err = os.Open(modelPath)
	if err != nil {
		loadedFiles.Close()
		return nil, nil, fmt.Errorf("failed to open model file: %w", err)
	}

	// Open shortlist file
	loadedFiles.ShortlistFile, err = os.Open(shortlistPath)
	if err != nil {
		loadedFiles.Close()
		return nil, nil, fmt.Errorf("failed to open shortlist file: %w", err)
	}

	// Open vocabulary files
	vocabularies := make([]io.Reader, len(vocabPaths))
	for i, vocabPath := range vocabPaths {
		vocabFile, err := os.Open(vocabPath)
		if err != nil {
			loadedFiles.Close()
			return nil, nil, fmt.Errorf("failed to open vocabulary file %s: %w", vocabPath, err)
		}
		loadedFiles.VocabFiles = append(loadedFiles.VocabFiles, vocabFile)
		vocabularies[i] = vocabFile
	}

	// Create engine config
	engineCfg := engine.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: nil,
			Stdout: nil,
		},
		FilesBundle: engine.FilesBundle{
			Model:            loadedFiles.ModelFile,
			LexicalShortlist: loadedFiles.ShortlistFile,
			Vocabularies:     vocabularies,
		},
		CacheSize:       1024,
		BergamotOptions: engine.DefaultBergamotOptions(),
	}

	// Create translator
	translator, err := engine.New(ctx, engineCfg)
	if err != nil {
		loadedFiles.Close()
		return nil, nil, fmt.Errorf("failed to create translator: %w", err)
	}

	return translator, loadedFiles, nil
}

// discoverModelFiles automatically finds model files in a directory
// Returns: modelPath, shortlistPath, vocabPaths, error
func discoverModelFiles(dir string) (string, string, []string, error) {
	var modelPath, shortlistPath string
	var vocabPaths []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Detect model file (.bin but not lex*.bin)
		if filepath.Ext(name) == ".bin" && !hasPrefix(name, "lex") {
			modelPath = fullPath
		}
		// Detect shortlist file (lex*.bin)
		if filepath.Ext(name) == ".bin" && hasPrefix(name, "lex") {
			shortlistPath = fullPath
		}
		// Detect vocabulary files (.spm)
		if filepath.Ext(name) == ".spm" {
			vocabPaths = append(vocabPaths, fullPath)
		}
	}

	if modelPath == "" {
		return "", "", nil, fmt.Errorf("model file (*.bin) not found in directory")
	}
	if shortlistPath == "" {
		return "", "", nil, fmt.Errorf("shortlist file (lex*.bin) not found in directory")
	}
	if len(vocabPaths) == 0 {
		return "", "", nil, fmt.Errorf("vocabulary files (*.spm) not found in directory")
	}

	return modelPath, shortlistPath, vocabPaths, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
