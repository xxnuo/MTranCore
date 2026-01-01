package MTranCore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

	// Cache size for translation
	CacheSize uint `json:"cache_size,omitempty"`

	// MaxLengthBreak defines the maximum text length (in characters) before auto-splitting
	// Default is 200 if not set.
	MaxLengthBreak int `json:"max_length_break,omitempty"`
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
// This is the shared logic used by both CLI and server applications
func CreateTranslator(ctx context.Context, cfg EngineConfig) (*Translator, *LoadedFiles, error) {
	var modelPath, shortlistPath string
	var vocabPaths []string
	var err error

	// Auto-discover model files if ModelDir is provided
	if cfg.ModelDir != "" {
		modelPath, shortlistPath, vocabPaths, err = DiscoverModelFiles(cfg.ModelDir)
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

	// Set default cache size if not specified
	cacheSize := cfg.CacheSize
	if cacheSize == 0 {
		cacheSize = 1024
	}

	// Create engine config
	engineCfg := Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: nil,
			Stdout: nil,
		},
		FilesBundle: FilesBundle{
			Model:            loadedFiles.ModelFile,
			LexicalShortlist: loadedFiles.ShortlistFile,
			Vocabularies:     vocabularies,
		},
		CacheSize:       cacheSize,
		BergamotOptions: DefaultBergamotOptions(),
		MaxLengthBreak:  cfg.MaxLengthBreak,
	}

	// Create translator
	translator, err := New(ctx, engineCfg)
	if err != nil {
		loadedFiles.Close()
		return nil, nil, fmt.Errorf("failed to create translator: %w", err)
	}

	return translator, loadedFiles, nil
}

// DiscoverModelFiles automatically finds model files in a directory
// Returns: modelPath, shortlistPath, vocabPaths, error
func DiscoverModelFiles(dir string) (string, string, []string, error) {
	var modelPath, shortlistPath string
	var vocabPaths []string
	var srcVocab, trgVocab, sharedVocab string

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
			nameLower := strings.ToLower(name)
			if strings.Contains(nameLower, "srcvocab") || strings.Contains(nameLower, "source") {
				srcVocab = fullPath
			} else if strings.Contains(nameLower, "trgvocab") || strings.Contains(nameLower, "target") {
				trgVocab = fullPath
			} else if strings.Contains(nameLower, "vocab") {
				sharedVocab = fullPath
			}
		}
	}

	if modelPath == "" {
		return "", "", nil, fmt.Errorf("model file (*.bin) not found in directory")
	}
	if shortlistPath == "" {
		return "", "", nil, fmt.Errorf("shortlist file (lex*.bin) not found in directory")
	}

	// Determine which vocabularies to use:
	// 1. If shared vocab exists, prefer it (for tied-embeddings models)
	// 2. Otherwise, use separate srcvocab and trgvocab if both exist
	if sharedVocab != "" {
		vocabPaths = []string{sharedVocab}
	} else if srcVocab != "" && trgVocab != "" {
		vocabPaths = []string{srcVocab, trgVocab}
	} else {
		return "", "", nil, fmt.Errorf("vocabulary files not found in directory (need either shared vocab or srcvocab+trgvocab)")
	}

	return modelPath, shortlistPath, vocabPaths, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
