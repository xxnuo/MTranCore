package MTranCore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	engine "github.com/xxnuo/MTranCore/engine"
)

// TestDiscoverModelFiles tests the model file discovery functionality
func TestDiscoverModelFiles(t *testing.T) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	modelDir := filepath.Join(projectRoot, "models", "enzh")

	// Check if model directory exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		t.Skipf("Model directory not found: %s", modelDir)
	}

	t.Run("ValidModelDirectory", func(t *testing.T) {
		modelPath, shortlistPath, vocabPaths, err := engine.DiscoverModelFiles(modelDir)
		if err != nil {
			t.Fatalf("DiscoverModelFiles failed: %v", err)
		}

		// Verify model path is not empty and file exists
		if modelPath == "" {
			t.Error("Expected non-empty model path")
		}
		if _, err := os.Stat(modelPath); err != nil {
			t.Errorf("Model file not found: %v", err)
		}

		// Verify shortlist path is not empty and file exists
		if shortlistPath == "" {
			t.Error("Expected non-empty shortlist path")
		}
		if _, err := os.Stat(shortlistPath); err != nil {
			t.Errorf("Shortlist file not found: %v", err)
		}

		// Verify vocab paths are not empty and files exist
		if len(vocabPaths) == 0 {
			t.Error("Expected at least one vocabulary path")
		}
		for i, vocabPath := range vocabPaths {
			if vocabPath == "" {
				t.Errorf("Expected non-empty vocabulary path at index %d", i)
			}
			if _, err := os.Stat(vocabPath); err != nil {
				t.Errorf("Vocabulary file not found at index %d: %v", i, err)
			}
		}
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		_, _, _, err := engine.DiscoverModelFiles("/nonexistent/directory")
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		// Create a temporary empty directory
		tmpDir := t.TempDir()

		_, _, _, err := engine.DiscoverModelFiles(tmpDir)
		if err == nil {
			t.Error("Expected error for empty directory")
		}
	})

	t.Run("DirectoryWithMissingFiles", func(t *testing.T) {
		// Create a temporary directory with only a model file
		tmpDir := t.TempDir()
		modelFile := filepath.Join(tmpDir, "model.bin")
		if err := os.WriteFile(modelFile, []byte("dummy"), 0644); err != nil {
			t.Fatalf("Failed to create dummy model file: %v", err)
		}

		_, _, _, err := engine.DiscoverModelFiles(tmpDir)
		if err == nil {
			t.Error("Expected error for directory missing required files")
		}
	})
}

// TestEngineConfig tests the EngineConfig struct and its validation
func TestEngineConfig(t *testing.T) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	modelDir := filepath.Join(projectRoot, "models", "enzh")

	// Check if model directory exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		t.Skipf("Model directory not found: %s", modelDir)
	}

	t.Run("ConfigWithModelDir", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelDir:  modelDir,
			CacheSize: 1024,
		}

		if cfg.ModelDir != modelDir {
			t.Errorf("Expected ModelDir to be %s, got %s", modelDir, cfg.ModelDir)
		}
		if cfg.CacheSize != 1024 {
			t.Errorf("Expected CacheSize to be 1024, got %d", cfg.CacheSize)
		}
	})

	t.Run("ConfigWithIndividualPaths", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelPath:            "/path/to/model.bin",
			LexicalShortlistPath: "/path/to/lex.bin",
			VocabularyPaths:      []string{"/path/to/vocab.spm"},
			CacheSize:            2048,
		}

		if cfg.ModelPath != "/path/to/model.bin" {
			t.Errorf("Expected ModelPath to be /path/to/model.bin, got %s", cfg.ModelPath)
		}
		if cfg.LexicalShortlistPath != "/path/to/lex.bin" {
			t.Errorf("Expected LexicalShortlistPath to be /path/to/lex.bin, got %s", cfg.LexicalShortlistPath)
		}
		if len(cfg.VocabularyPaths) != 1 {
			t.Errorf("Expected 1 VocabularyPath, got %d", len(cfg.VocabularyPaths))
		}
	})
}

// TestCreateTranslator tests the CreateTranslator function
func TestCreateTranslator(t *testing.T) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}

	modelDir := filepath.Join(projectRoot, "models", "enzh")

	// Check if model directory exists
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		t.Skipf("Model directory not found: %s", modelDir)
	}

	ctx := context.Background()

	t.Run("CreateWithModelDir", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelDir:  modelDir,
			CacheSize: 1024,
		}

		translator, loadedFiles, err := engine.CreateTranslator(ctx, cfg)
		if err != nil {
			t.Fatalf("CreateTranslator failed: %v", err)
		}
		defer func() {
			if translator != nil {
				translator.Close(ctx)
			}
			if loadedFiles != nil {
				loadedFiles.Close()
			}
		}()

		if translator == nil {
			t.Error("Expected non-nil translator")
		}
		if loadedFiles == nil {
			t.Error("Expected non-nil loadedFiles")
		}

		// Verify loaded files are properly opened
		if loadedFiles.ModelFile == nil {
			t.Error("Expected non-nil ModelFile")
		}
		if loadedFiles.ShortlistFile == nil {
			t.Error("Expected non-nil ShortlistFile")
		}
		if len(loadedFiles.VocabFiles) == 0 {
			t.Error("Expected at least one VocabFile")
		}

		// Test translation to verify the translator works
		req := engine.TranslationRequest{
			Text: "Hello, world!",
			Options: engine.TranslationOptions{
				HTML: false,
			},
		}

		result, err := translator.Translate(ctx, req)
		if err != nil {
			t.Errorf("Translation failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty translation result")
		}
	})

	t.Run("CreateWithMissingModelDir", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelDir:  "/nonexistent/directory",
			CacheSize: 1024,
		}

		translator, loadedFiles, err := engine.CreateTranslator(ctx, cfg)
		if err == nil {
			if translator != nil {
				translator.Close(ctx)
			}
			if loadedFiles != nil {
				loadedFiles.Close()
			}
			t.Error("Expected error for non-existent model directory")
		}
	})

	t.Run("CreateWithInvalidIndividualPaths", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelPath:            "/nonexistent/model.bin",
			LexicalShortlistPath: "/nonexistent/lex.bin",
			VocabularyPaths:      []string{"/nonexistent/vocab.spm"},
			CacheSize:            1024,
		}

		translator, loadedFiles, err := engine.CreateTranslator(ctx, cfg)
		if err == nil {
			if translator != nil {
				translator.Close(ctx)
			}
			if loadedFiles != nil {
				loadedFiles.Close()
			}
			t.Error("Expected error for non-existent model files")
		}
	})

	t.Run("CreateWithMissingRequiredFields", func(t *testing.T) {
		// Missing all required fields
		cfg := engine.EngineConfig{
			CacheSize: 1024,
		}

		translator, loadedFiles, err := engine.CreateTranslator(ctx, cfg)
		if err == nil {
			if translator != nil {
				translator.Close(ctx)
			}
			if loadedFiles != nil {
				loadedFiles.Close()
			}
			t.Error("Expected error for missing required fields")
		}
	})

	t.Run("CreateWithDefaultCacheSize", func(t *testing.T) {
		cfg := engine.EngineConfig{
			ModelDir: modelDir,
			// CacheSize not set, should use default
		}

		translator, loadedFiles, err := engine.CreateTranslator(ctx, cfg)
		if err != nil {
			t.Fatalf("CreateTranslator failed: %v", err)
		}
		defer func() {
			if translator != nil {
				translator.Close(ctx)
			}
			if loadedFiles != nil {
				loadedFiles.Close()
			}
		}()

		if translator == nil {
			t.Error("Expected non-nil translator with default cache size")
		}
	})
}

// TestLoadedFilesClose tests the LoadedFiles.Close method
func TestLoadedFilesClose(t *testing.T) {
	// Create temporary files for testing
	tmpDir := t.TempDir()

	modelFile, err := os.Create(filepath.Join(tmpDir, "model.bin"))
	if err != nil {
		t.Fatalf("Failed to create temp model file: %v", err)
	}

	shortlistFile, err := os.Create(filepath.Join(tmpDir, "lex.bin"))
	if err != nil {
		t.Fatalf("Failed to create temp shortlist file: %v", err)
	}

	vocabFile1, err := os.Create(filepath.Join(tmpDir, "vocab1.spm"))
	if err != nil {
		t.Fatalf("Failed to create temp vocab file 1: %v", err)
	}

	vocabFile2, err := os.Create(filepath.Join(tmpDir, "vocab2.spm"))
	if err != nil {
		t.Fatalf("Failed to create temp vocab file 2: %v", err)
	}

	// Create LoadedFiles instance
	lf := &engine.LoadedFiles{
		ModelFile:     modelFile,
		ShortlistFile: shortlistFile,
		VocabFiles:    []*os.File{vocabFile1, vocabFile2},
	}

	// Close all files
	lf.Close()

	// Verify all files are closed by trying to read from them
	// (Reading from a closed file should fail)
	buf := make([]byte, 1)
	if _, err := modelFile.Read(buf); err == nil {
		t.Error("Expected error when reading from closed model file")
	}
	if _, err := shortlistFile.Read(buf); err == nil {
		t.Error("Expected error when reading from closed shortlist file")
	}
	if _, err := vocabFile1.Read(buf); err == nil {
		t.Error("Expected error when reading from closed vocab file 1")
	}
	if _, err := vocabFile2.Read(buf); err == nil {
		t.Error("Expected error when reading from closed vocab file 2")
	}
}

// TestLoadedFilesCloseWithNilFiles tests that Close handles nil files gracefully
func TestLoadedFilesCloseWithNilFiles(t *testing.T) {
	lf := &engine.LoadedFiles{
		ModelFile:     nil,
		ShortlistFile: nil,
		VocabFiles:    nil,
	}

	// Should not panic
	lf.Close()
}

// TestLoadedFilesCloseWithPartialFiles tests Close with some nil files
func TestLoadedFilesCloseWithPartialFiles(t *testing.T) {
	tmpDir := t.TempDir()

	modelFile, err := os.Create(filepath.Join(tmpDir, "model.bin"))
	if err != nil {
		t.Fatalf("Failed to create temp model file: %v", err)
	}

	lf := &engine.LoadedFiles{
		ModelFile:     modelFile,
		ShortlistFile: nil,
		VocabFiles:    []*os.File{nil},
	}

	// Should not panic
	lf.Close()

	// Verify model file is closed
	buf := make([]byte, 1)
	if _, err := modelFile.Read(buf); err == nil {
		t.Error("Expected error when reading from closed model file")
	}
}
