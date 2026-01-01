package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"io"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServer_Trans(t *testing.T) {
	cfg := &Config{WorkDir: "./", EnableHTTP: true}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	defer server.Close()
	t.Run("engine not health", func(t *testing.T) {
		reqBody := TransRequest{
			Text: "Hello, world!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/trans", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusServiceUnavailable {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusServiceUnavailable)
		}
		var result StandardResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if result.Code != CodeTransInvalidParams {
			t.Errorf("Code = %d, want %d", result.Code, CodeTransInvalidParams)
		}
	})
	t.Run("empty text", func(t *testing.T) {
		reqBody := TransRequest{
			Text: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/trans", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
		}
		var result StandardResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if result.Code != CodeTransInvalidParams {
			t.Errorf("Code = %d, want %d", result.Code, CodeTransInvalidParams)
		}
	})
	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/trans", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
		}
	})
}
func TestServer_Trans_Success(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}
	modelDir := filepath.Dir(modelPath)
	cfg := &Config{WorkDir: "./", EnableHTTP: true}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	defer server.Close()
	// Load engine first
	loadReqBody := LoadOptions{
		Path: modelDir,
	}
	ctx := context.Background()
	result := em.Load(ctx, loadReqBody)
	if !result.Success {
		t.Fatalf("Failed to load engine: %s", result.ErrorMessage)
	}
	// Test translation
	t.Run("translate text", func(t *testing.T) {
		reqBody := TransRequest{
			Text: "Hello, world!",
			HTML: false,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/trans", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Status code = %d, want %d, body: %s", resp.StatusCode, fiber.StatusOK, string(bodyBytes))
		}
		// Response should be plain text now
		translatedText, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}
		if string(translatedText) == "" {
			t.Error("translated_text is empty")
		}
		t.Logf("Translation: 'Hello, world!' -> '%s'", string(translatedText))
	})
	t.Run("translate HTML", func(t *testing.T) {
		reqBody := TransRequest{
			Text: "<p>Hello, world!</p>",
			HTML: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/trans", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Status code = %d, want %d, body: %s", resp.StatusCode, fiber.StatusOK, string(bodyBytes))
		}
		// Response should be plain text now
		translatedText, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}
		if string(translatedText) == "" {
			t.Error("translated_text is empty")
		}
		t.Logf("HTML Translation: '<p>Hello, world!</p>' -> '%s'", string(translatedText))
	})
}
func TestServer_ConcurrentTranslations(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}
	modelDir := filepath.Dir(modelPath)
	cfg := &Config{WorkDir: "./", EnableHTTP: true}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	defer server.Close()
	// Load engine
	loadReqBody := LoadOptions{
		Path: modelDir,
	}
	ctx := context.Background()
	result := em.Load(ctx, loadReqBody)
	if !result.Success {
		t.Fatalf("Failed to load engine: %s", result.ErrorMessage)
	}
	// Test concurrent translations
	testTexts := []string{
		"Hello!",
		"Good morning!",
		"How are you?",
		"Thank you!",
		"Goodbye!",
	}
	var wg sync.WaitGroup
	results := make([]string, len(testTexts))
	errors := make([]error, len(testTexts))
	for i, text := range testTexts {
		wg.Add(1)
		go func(idx int, txt string) {
			defer wg.Done()
			reqBody := TransRequest{Text: txt}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/trans", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
			if err != nil {
				errors[idx] = err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusOK {
				errors[idx] = fmt.Errorf("status code = %d", resp.StatusCode)
				return
			}
			// Response should be plain text now
			translatedText, err := io.ReadAll(resp.Body)
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = string(translatedText)
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
		t.Logf("Concurrent[%d] Input: %s -> Output: %s", i, text, results[i])
	}
}
