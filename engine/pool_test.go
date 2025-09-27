package MTranCore_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	engine "github.com/xxnuo/MTranCore/engine"
	"github.com/xxnuo/MTranCore/internal/wasm"
)

// Mock reader for testing error cases
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

// Mock reader that returns specific data
type mockReader struct {
	data []byte
	pos  int
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func newMockReader(data string) *mockReader {
	return &mockReader{data: []byte(data), pos: 0}
}

func TestPoolConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  engine.PoolConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: engine.PoolConfig{
				Config: engine.Config{
					FilesBundle: engine.FilesBundle{
						Model:            strings.NewReader("model"),
						LexicalShortlist: strings.NewReader("shortlist"),
						Vocabularies:     []io.Reader{strings.NewReader("vocab")},
					},
				},
				PoolSize: 2,
			},
			wantErr: false,
		},
		{
			name: "zero pool size",
			config: engine.PoolConfig{
				Config: engine.Config{
					FilesBundle: engine.FilesBundle{
						Model:            strings.NewReader("model"),
						LexicalShortlist: strings.NewReader("shortlist"),
						Vocabularies:     []io.Reader{strings.NewReader("vocab")},
					},
				},
				PoolSize: 0,
			},
			wantErr: true,
			errMsg:  "zero pool size",
		},
		{
			name: "missing model",
			config: engine.PoolConfig{
				Config: engine.Config{
					FilesBundle: engine.FilesBundle{
						LexicalShortlist: strings.NewReader("shortlist"),
						Vocabularies:     []io.Reader{strings.NewReader("vocab")},
					},
				},
				PoolSize: 2,
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "missing shortlist",
			config: engine.PoolConfig{
				Config: engine.Config{
					FilesBundle: engine.FilesBundle{
						Model:        strings.NewReader("model"),
						Vocabularies: []io.Reader{strings.NewReader("vocab")},
					},
				},
				PoolSize: 2,
			},
			wantErr: true,
			errMsg:  "lexical shortlist is required",
		},
		{
			name: "missing vocabularies",
			config: engine.PoolConfig{
				Config: engine.Config{
					FilesBundle: engine.FilesBundle{
						Model:            strings.NewReader("model"),
						LexicalShortlist: strings.NewReader("shortlist"),
					},
				},
				PoolSize: 2,
			},
			wantErr: true,
			errMsg:  "at least one vocabulary is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("PoolConfig.Validate() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PoolConfig.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PoolConfig.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestNewPool_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid config", func(t *testing.T) {
		cfg := engine.PoolConfig{
			PoolSize: 0, // Invalid
		}
		_, err := engine.NewPool(ctx, cfg)
		if err == nil {
			t.Fatal("NewPool() expected error for invalid config, got nil")
		}
		if !strings.Contains(err.Error(), "zero pool size") {
			t.Errorf("NewPool() error = %v, want error containing 'zero pool size'", err)
		}
	})

	t.Run("file read error", func(t *testing.T) {
		cfg := engine.PoolConfig{
			Config: engine.Config{
				FilesBundle: engine.FilesBundle{
					Model:            &errorReader{err: errors.New("model read error")},
					LexicalShortlist: strings.NewReader("shortlist"),
					Vocabularies:     []io.Reader{strings.NewReader("vocab")},
				},
			},
			PoolSize: 1,
		}
		_, err := engine.NewPool(ctx, cfg)
		if err == nil {
			t.Fatal("NewPool() expected error for file read error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read model") {
			t.Errorf("NewPool() error = %v, want error containing 'failed to read model'", err)
		}
	})

	t.Run("shortlist read error", func(t *testing.T) {
		cfg := engine.PoolConfig{
			Config: engine.Config{
				FilesBundle: engine.FilesBundle{
					Model:            strings.NewReader("model"),
					LexicalShortlist: &errorReader{err: errors.New("shortlist read error")},
					Vocabularies:     []io.Reader{strings.NewReader("vocab")},
				},
			},
			PoolSize: 1,
		}
		_, err := engine.NewPool(ctx, cfg)
		if err == nil {
			t.Fatal("NewPool() expected error for shortlist read error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read shortlist") {
			t.Errorf("NewPool() error = %v, want error containing 'failed to read shortlist'", err)
		}
	})

	t.Run("vocabulary read error", func(t *testing.T) {
		cfg := engine.PoolConfig{
			Config: engine.Config{
				FilesBundle: engine.FilesBundle{
					Model:            strings.NewReader("model"),
					LexicalShortlist: strings.NewReader("shortlist"),
					Vocabularies:     []io.Reader{&errorReader{err: errors.New("vocab read error")}},
				},
			},
			PoolSize: 1,
		}
		_, err := engine.NewPool(ctx, cfg)
		if err == nil {
			t.Fatal("NewPool() expected error for vocabulary read error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read vocabulary") {
			t.Errorf("NewPool() error = %v, want error containing 'failed to read vocabulary'", err)
		}
	})
}

func TestNewPool_DefaultOptions(t *testing.T) {
	// This test verifies that default bergamot options are set when nil
	// We can't actually create a real pool without proper model files,
	// but we can test the configuration validation path
	cfg := engine.PoolConfig{
		Config: engine.Config{
			FilesBundle: engine.FilesBundle{
				Model:            strings.NewReader("model"),
				LexicalShortlist: strings.NewReader("shortlist"),
				Vocabularies:     []io.Reader{strings.NewReader("vocab")},
			},
			BergamotOptions: nil, // Should be set to defaults
		},
		PoolSize: 1,
	}

	// Validate the config (this should work without creating actual translators)
	err := cfg.Validate()
	if err != nil {
		t.Errorf("PoolConfig.Validate() unexpected error = %v", err)
	}
}

// Helper function to create a mock pool for testing (without real WASM)
func createMockPoolConfig() engine.PoolConfig {
	return engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: engine.FilesBundle{
				Model:            strings.NewReader("mock model data"),
				LexicalShortlist: strings.NewReader("mock shortlist data"),
				Vocabularies:     []io.Reader{strings.NewReader("mock vocab data")},
			},
		},
		PoolSize: 2,
	}
}

func TestPool_ConcurrentTranslation(t *testing.T) {
	// Skip this test if models are not available (requires real model files)
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 3,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	// Test concurrent translations
	testTexts := []string{
		"Hello, World!",
		"Good morning!",
		"How are you today?",
		"This is a test message.",
		"Goodbye and have a nice day!",
	}

	var wg sync.WaitGroup
	results := make([]string, len(testTexts))
	errors := make([]error, len(testTexts))

	for i, text := range testTexts {
		wg.Add(1)
		go func(idx int, txt string) {
			defer wg.Done()
			result, err := pool.Translate(ctx, engine.TranslationRequest{Text: txt})
			results[idx] = result
			errors[idx] = err
		}(i, text)
	}

	wg.Wait()

	// Check results
	for i, text := range testTexts {
		if errors[i] != nil {
			t.Errorf("Translation %d error = %v for text: %s", i, errors[i], text)
		}
		if results[i] == "" {
			t.Errorf("Translation %d result is empty for text: %s", i, text)
		}
		t.Logf("Translation[%d] Input: %s -> Output: %s", i, text, results[i])
	}
}

func TestPool_TranslateMultiple(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 2,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	requests := []engine.TranslationRequest{
		{Text: "Hello"},
		{Text: "World"},
		{Text: "How are you?"},
	}

	results, err := pool.TranslateMultiple(ctx, requests...)
	if err != nil {
		t.Fatalf("TranslateMultiple() error = %v", err)
	}

	if len(results) != len(requests) {
		t.Errorf("TranslateMultiple() returned %d results, want %d", len(results), len(requests))
	}

	for i, result := range results {
		if result == "" {
			t.Errorf("TranslateMultiple() result[%d] is empty", i)
		}
		t.Logf("TranslateMultiple[%d] Input: %s -> Output: %s", i, requests[i].Text, result)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 1,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err = pool.Translate(cancelCtx, engine.TranslationRequest{Text: "Hello"})
	if err == nil {
		t.Error("Translate() expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Translate() error = %v, want error containing 'context canceled'", err)
	}
}

func TestPool_ContextTimeout(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 1,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	// Test context timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout

	_, err = pool.Translate(timeoutCtx, engine.TranslationRequest{Text: "Hello"})
	if err == nil {
		t.Error("Translate() expected error for timed out context, got nil")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Translate() error = %v, want timeout related error", err)
	}
}

func TestPool_Close(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 2,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	// Close the pool
	err = pool.Close(ctx)
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Try to use pool after closing
	_, err = pool.Translate(ctx, engine.TranslationRequest{Text: "Hello"})
	if err == nil {
		t.Error("Translate() after Close() expected error, got nil")
	}
	if !errors.Is(err, engine.ErrClosed) && !strings.Contains(err.Error(), "pool closed") {
		t.Errorf("Translate() after Close() error = %v, want ErrClosed", err)
	}
}

func TestPool_CloseTimeout(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 1,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}

	// Test close with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure timeout

	err = pool.Close(timeoutCtx)
	if err == nil {
		t.Error("Close() with timeout expected error, got nil")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Close() with timeout error = %v, want timeout related error", err)
	}

	// Note: Don't call Close again as the pool may have been partially closed
	// The timeout test verifies timeout behavior, not successful cleanup
}

func TestPool_SingleTranslation(t *testing.T) {
	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 1,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	// Test single translation
	text := "Hello, world!"
	result, err := pool.Translate(ctx, engine.TranslationRequest{
		Text: text,
		Options: engine.TranslationOptions{
			HTML: false,
		},
	})
	if err != nil {
		t.Fatalf("Translate() error = %v", err)
	}
	if result == "" {
		t.Error("Translate() result is empty")
	}
	t.Logf("Single translation: %s -> %s", text, result)
}

func TestPool_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Skip this test if models are not available
	bundle, err := getTestBundle()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	ctx := context.Background()
	cfg := engine.PoolConfig{
		Config: engine.Config{
			CompileConfig: wasm.CompileConfig{
				Stderr: io.Discard,
				Stdout: io.Discard,
			},
			FilesBundle: bundle,
		},
		PoolSize: 3,
	}

	pool, err := engine.NewPool(ctx, cfg)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close(ctx)

	// Stress test with many concurrent translations
	numGoroutines := 20
	translationsPerGoroutine := 10

	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*translationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < translationsPerGoroutine; j++ {
				text := fmt.Sprintf("Worker %d, message %d: Hello world!", workerID, j)
				_, err := pool.Translate(ctx, engine.TranslationRequest{Text: text})
				if err != nil {
					errChan <- fmt.Errorf("worker %d, message %d: %w", workerID, j, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		t.Errorf("Stress test had %d errors:", len(errs))
		for i, err := range errs {
			if i < 5 { // Show first 5 errors
				t.Errorf("  Error %d: %v", i+1, err)
			}
		}
		if len(errs) > 5 {
			t.Errorf("  ... and %d more errors", len(errs)-5)
		}
	}

	t.Logf("Stress test completed: %d goroutines × %d translations = %d total translations",
		numGoroutines, translationsPerGoroutine, numGoroutines*translationsPerGoroutine)
}
