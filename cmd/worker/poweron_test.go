package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestServer_Poweron(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	t.Run("missing path", func(t *testing.T) {
		reqBody := PoweronRequest{}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
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

		if result.Code != CodePoweronInvalidParams {
			t.Errorf("Code = %d, want %d", result.Code, CodePoweronInvalidParams)
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		reqBody := PoweronRequest{
			Path: "/non/existent/path",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
		}

		var result StandardResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Code != CodePoweronPathNotExists {
			t.Errorf("Code = %d, want %d", result.Code, CodePoweronPathNotExists)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/poweron", strings.NewReader("invalid json"))
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

func TestServer_Poweron_Success(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	// Get the model directory
	modelDir := filepath.Dir(modelPath)

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	reqBody := PoweronRequest{
		Path: modelDir,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 30 * time.Second}) // 30 second timeout for engine loading
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

	// Verify ready shows engine loaded
	readyReq := httptest.NewRequest("GET", "/ready", nil)
	readyResp, err := server.GetApp().Test(readyReq)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer readyResp.Body.Close()

	var readyResult StandardResponse
	if err := json.NewDecoder(readyResp.Body).Decode(&readyResult); err != nil {
		t.Fatalf("Failed to decode ready response: %v", err)
	}

	readyData, _ := readyResult.Data.(map[string]interface{})
	if ready, ok := readyData["ready"].(bool); !ok || !ready {
		t.Errorf("ready = %v, want true", readyData["ready"])
	}
}

func TestServer_Poweron_IndividualPaths(t *testing.T) {
	modelPath, shortlistPath, vocabPaths, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	// Test with individual file paths
	reqBody := PoweronRequest{
		ModelPath:            modelPath,
		LexicalShortlistPath: shortlistPath,
		VocabularyPaths:      vocabPaths,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
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

	// Verify ready shows engine loaded
	readyReq := httptest.NewRequest("GET", "/ready", nil)
	readyResp, err := server.GetApp().Test(readyReq)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer readyResp.Body.Close()

	var readyResult StandardResponse
	if err := json.NewDecoder(readyResp.Body).Decode(&readyResult); err != nil {
		t.Fatalf("Failed to decode ready response: %v", err)
	}

	readyData, _ := readyResult.Data.(map[string]interface{})
	if ready, ok := readyData["ready"].(bool); !ok || !ready {
		t.Errorf("ready = %v, want true", readyData["ready"])
	}
}

func TestServer_Poweron_VocabularyPathMerge(t *testing.T) {
	modelPath, shortlistPath, vocabPaths, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	if len(vocabPaths) < 2 {
		t.Skip("Test requires at least 2 vocabulary files")
	}

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	// Test vocabulary_path and vocabulary_paths merging
	reqBody := PoweronRequest{
		ModelPath:            modelPath,
		LexicalShortlistPath: shortlistPath,
		VocabularyPath:       vocabPaths[0],
		VocabularyPaths:      vocabPaths[1:],
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.GetApp().Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
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
}

func TestServer_ReloadEngine(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	modelDir := filepath.Dir(modelPath)

	cfg := &Config{WorkDir: "./"}
	server := NewServer(cfg)
	defer server.Close()

	loadReqBody := PoweronRequest{
		Path: modelDir,
	}

	// Load engine first time
	body1, _ := json.Marshal(loadReqBody)
	req1 := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := server.GetApp().Test(req1, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != fiber.StatusOK {
		t.Fatalf("First load status code = %d, want %d", resp1.StatusCode, fiber.StatusOK)
	}

	// Load engine second time (should replace existing engine)
	body2, _ := json.Marshal(loadReqBody)
	req2 := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := server.GetApp().Test(req2, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != fiber.StatusOK {
		bodyBytes, _ := io.ReadAll(resp2.Body)
		t.Fatalf("Second load status code = %d, want %d, body: %s", resp2.StatusCode, fiber.StatusOK, string(bodyBytes))
	}

	var result StandardResponse
	bodyBytes, _ := io.ReadAll(resp2.Body)
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Code != CodeSuccess {
		t.Errorf("Code = %d, want %d", result.Code, CodeSuccess)
	}
}
