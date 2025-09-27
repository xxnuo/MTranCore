package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestServer_Compute(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	t.Run("engine not ready", func(t *testing.T) {
		reqBody := ComputeRequest{
			Text: "Hello, world!",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/compute", bytes.NewReader(body))
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

		if result.Code != CodeComputeInvalidParams {
			t.Errorf("Code = %d, want %d", result.Code, CodeComputeInvalidParams)
		}
	})

	t.Run("empty text", func(t *testing.T) {
		reqBody := ComputeRequest{
			Text: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/compute", bytes.NewReader(body))
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

		if result.Code != CodeComputeInvalidParams {
			t.Errorf("Code = %d, want %d", result.Code, CodeComputeInvalidParams)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/compute", strings.NewReader("invalid json"))
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

func TestServer_Compute_Success(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	modelDir := filepath.Dir(modelPath)

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	// Load engine first
	loadReqBody := PoweronRequest{
		Path: modelDir,
	}
	loadBody, _ := json.Marshal(loadReqBody)
	loadReq := httptest.NewRequest("POST", "/poweron", bytes.NewReader(loadBody))
	loadReq.Header.Set("Content-Type", "application/json")

	loadResp, err := server.GetApp().Test(loadReq, fiber.TestConfig{Timeout: 30 * time.Second}) // 30 second timeout
	if err != nil {
		t.Fatalf("Failed to load engine: %v", err)
	}
	loadResp.Body.Close()

	if loadResp.StatusCode != fiber.StatusOK {
		t.Fatalf("Failed to load engine, status code = %d", loadResp.StatusCode)
	}

	// Test translation
	t.Run("translate text", func(t *testing.T) {
		reqBody := ComputeRequest{
			Text: "Hello, world!",
			HTML: false,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/compute", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Status code = %d, want %d, body: %s", resp.StatusCode, fiber.StatusOK, string(bodyBytes))
		}

		var result StandardResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Code != CodeSuccess {
			t.Errorf("Code = %d, want %d", result.Code, CodeSuccess)
		}

		computeData, _ := result.Data.(map[string]interface{})
		translatedText := computeData["translated_text"].(string)

		if translatedText == "" {
			t.Error("translated_text is empty")
		}
		t.Logf("Translation: 'Hello, world!' -> '%s'", translatedText)
	})

	t.Run("translate HTML", func(t *testing.T) {
		reqBody := ComputeRequest{
			Text: "<p>Hello, world!</p>",
			HTML: true,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/compute", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("Status code = %d, want %d, body: %s", resp.StatusCode, fiber.StatusOK, string(bodyBytes))
		}

		var result StandardResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		computeData, _ := result.Data.(map[string]interface{})
		translatedText := computeData["translated_text"].(string)

		if translatedText == "" {
			t.Error("translated_text is empty")
		}
		t.Logf("HTML Translation: '<p>Hello, world!</p>' -> '%s'", translatedText)
	})
}

func TestServer_ConcurrentTranslations(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	modelDir := filepath.Dir(modelPath)

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	// Load engine
	loadReqBody := PoweronRequest{
		Path: modelDir,
	}
	loadBody, _ := json.Marshal(loadReqBody)
	loadReq := httptest.NewRequest("POST", "/poweron", bytes.NewReader(loadBody))
	loadReq.Header.Set("Content-Type", "application/json")
	loadResp, err := server.GetApp().Test(loadReq, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("Failed to load engine: %v", err)
	}
	loadResp.Body.Close()

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

			reqBody := ComputeRequest{Text: txt}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/compute", bytes.NewReader(body))
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

			var result StandardResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				errors[idx] = err
				return
			}

			computeData, _ := result.Data.(map[string]interface{})
			results[idx] = computeData["translated_text"].(string)
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
