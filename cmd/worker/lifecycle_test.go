package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestServer_Close(t *testing.T) {
	t.Run("close without engine", func(t *testing.T) {
		cfg := &Config{WorkDir: "./"}
		server := NewUnifiedServer(cfg)
		err := server.Close()
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("close with engine", func(t *testing.T) {
		modelPath, _, _, err := getTestModelPaths()
		if err != nil {
			t.Skipf("Skipping test, model files not available: %v", err)
		}

		modelDir := filepath.Dir(modelPath)

		cfg := &Config{WorkDir: "./"}
		server := NewUnifiedServer(cfg)

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

		// Close server
		err = server.Close()
		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})
}

func TestServer_ThreadSafety(t *testing.T) {
	modelPath, _, _, err := getTestModelPaths()
	if err != nil {
		t.Skipf("Skipping test, model files not available: %v", err)
	}

	modelDir := filepath.Dir(modelPath)

	cfg := &Config{WorkDir: "./"}
	server := NewUnifiedServer(cfg)
	defer server.Close()

	var wg sync.WaitGroup
	iterations := 10

	// Concurrent ready checks
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/ready", nil)
			resp, _ := server.GetApp().Test(req)
			if resp != nil {
				resp.Body.Close()
			}
		}()
	}

	// Concurrent load operations
	loadReqBody := PoweronRequest{
		Path: modelDir,
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, _ := json.Marshal(loadReqBody)
			req := httptest.NewRequest("POST", "/poweron", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := server.GetApp().Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
			if resp != nil {
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	t.Log("Thread safety test completed successfully")
}
