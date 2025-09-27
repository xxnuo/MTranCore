package MTranCore_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/wasm"
)

func getProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Check if we're already in the project root
	if _, err := os.Stat(filepath.Join(cwd, "models")); err == nil {
		return cwd, nil
	}

	// Search upward for the models directory
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "models")); err == nil {
			return dir, nil
		}
	}

	return "", errors.New("cannot find project root directory containing models folder")
}

func getTestBundle() (engine.FilesBundle, error) {
	projectRoot, err := getProjectRoot()
	if err != nil {
		return engine.FilesBundle{}, err
	}

	// Open model files
	modelPath := filepath.Join(projectRoot, "models", "enzh", "model.enzh.intgemm.alphas.bin")
	modelFile, err := os.Open(modelPath)
	if err != nil {
		return engine.FilesBundle{}, fmt.Errorf("failed to open model file: %w", err)
	}

	shortlistPath := filepath.Join(projectRoot, "models", "enzh", "lex.50.50.enzh.s2t.bin")
	shortlistFile, err := os.Open(shortlistPath)
	if err != nil {
		modelFile.Close()
		return engine.FilesBundle{}, fmt.Errorf("failed to open shortlist file: %w", err)
	}

	srcVocabPath := filepath.Join(projectRoot, "models", "enzh", "srcvocab.enzh.spm")
	srcVocabFile, err := os.Open(srcVocabPath)
	if err != nil {
		modelFile.Close()
		shortlistFile.Close()
		return engine.FilesBundle{}, fmt.Errorf("failed to open src vocab file: %w", err)
	}

	trgVocabPath := filepath.Join(projectRoot, "models", "enzh", "trgvocab.enzh.spm")
	trgVocabFile, err := os.Open(trgVocabPath)
	if err != nil {
		modelFile.Close()
		shortlistFile.Close()
		srcVocabFile.Close()
		return engine.FilesBundle{}, fmt.Errorf("failed to open trg vocab file: %w", err)
	}

	return engine.FilesBundle{
		Model:            modelFile,
		LexicalShortlist: shortlistFile,
		Vocabularies:     []io.Reader{srcVocabFile, trgVocabFile},
	}, nil
}

func TestTranslator_SharedEngine(t *testing.T) {
	ctx := context.Background()

	bundle, err := getTestBundle()
	if err != nil {
		t.Fatalf("failed to get test bundle: %v", err)
	}

	translator, err := engine.New(ctx, engine.Config{
		CompileConfig: wasm.CompileConfig{
			Stderr: io.Discard,
			Stdout: io.Discard,
		},
		FilesBundle: bundle,
	})
	if err != nil {
		t.Fatalf("failed to create translator: %v", err)
	}
	defer translator.Close(ctx)

	// Test texts for concurrent translation
	testTexts := []string{
		"Hello, World!",
		"Good morning!",
		"How are you today?",
		"This is a test message.",
		"Goodbye and have a nice day!",
		"Welcome to our application.",
		"Please enter your information.",
		"Thank you for your patience.",
	}

	// Test 1: Sequential translations using the same engine
	t.Run("sequential_translations", func(t *testing.T) {
		results := make([]string, len(testTexts))
		for i, text := range testTexts {
			output, err := translator.Translate(ctx, engine.TranslationRequest{
				Text: text,
			})
			if err != nil {
				t.Fatalf("Translate() error = %v for text: %s", err, text)
			}
			if output == "" {
				t.Errorf("expected non-empty translation output for text: %s", text)
			}
			results[i] = output
			t.Logf("Sequential[%d] Input: %s -> Output: %s", i, text, output)
			time.Sleep(1 * time.Second)
		}
	})
}
