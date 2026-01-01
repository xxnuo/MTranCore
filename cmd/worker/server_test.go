package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestNewUnifiedServer(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	if server == nil {
		t.Fatal("NewUnifiedServer() returned nil")
	}
	if server.GetApp() == nil {
		t.Error("NewUnifiedServer() app is nil")
	}
	server.Close()
}

func TestServer_Health(t *testing.T) {
	cfg := &Config{
		WorkDir:    "./",
		EnableHTTP: true,
	}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	defer server.Close()

	t.Run("engine not health", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, err := server.GetApp().Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusOK)
		}

		var result StandardResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Code != CodeSuccess {
			t.Errorf("Code = %d, want %d", result.Code, CodeSuccess)
		}

		readyData, ok := result.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Data is not a map")
		}

		if health, ok := readyData["health"].(bool); !ok || health {
			t.Errorf("health = %v, want false", readyData["health"])
		}
	})
}

func TestServer_ErrorHandler(t *testing.T) {
	cfg := &Config{
		WorkDir:    "./",
		EnableHTTP: true,
	}
	em := &EngineManager{
		config: cfg,
		queue:  NewTranslationQueue(),
	}
	server := NewUnifiedServerWithEngine(cfg, em)
	defer server.Close()

	// Test with a non-existent route
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := server.GetApp().Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusNotFound)
	}
}
