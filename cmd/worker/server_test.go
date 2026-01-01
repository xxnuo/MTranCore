package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestNewUnifiedServer(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewUnifiedServer(cfg)
	if server == nil {
		t.Fatal("NewUnifiedServer() returned nil")
	}
	if server.GetApp() == nil {
		t.Error("NewUnifiedServer() app is nil")
	}
}

func TestServer_Health(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewUnifiedServer(cfg)
	defer server.Close()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := server.GetApp().Test(req)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Status code = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}
}

func TestServer_Ready(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewUnifiedServer(cfg)
	defer server.Close()

	t.Run("engine not ready", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
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

		if ready, ok := readyData["ready"].(bool); !ok || ready {
			t.Errorf("ready = %v, want false", readyData["ready"])
		}
	})
}

func TestServer_ErrorHandler(t *testing.T) {
	cfg := &Config{WorkDir: "./"}
	server := NewUnifiedServer(cfg)
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
