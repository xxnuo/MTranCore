package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"

	engine "github.com/xxnuo/MTranCore/engine"
)

// Server represents the HTTP server
type Server struct {
	app            *fiber.App
	translator     *engine.Translator
	loadedFiles    *LoadedFiles
	queue          *TranslationQueue // Request queue for sequential processing
	mu             sync.RWMutex
	shutdownCh     chan struct{}
	config         *Config
	activeReqs     int32 // atomic counter for active requests
	isShuttingDown atomic.Bool
}

// GetApp returns the fiber app instance (for testing)
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// PoweronRequest represents a poweron request
type PoweronRequest struct {
	Path string `json:"path"`
}

// PoweroffRequest represents a poweroff request
type PoweroffRequest struct {
	Time  int  `json:"time"`  // seconds to wait before shutdown
	Force bool `json:"force"` // force shutdown without waiting for requests
}

// ComputeRequest represents a compute (translation) request
type ComputeRequest struct {
	Text string `json:"text"`
	HTML bool   `json:"html,omitempty"`
}

// ComputeResponse represents the translation result
type ComputeResponse struct {
	TranslatedText string `json:"translated_text"`
}

// ReadyResponse represents the ready status response
type ReadyResponse struct {
	Ready bool `json:"ready"`
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *Config) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(NewErrorResponse(CodePoweronUnknownError, err.Error()))
		},
	})

	server := &Server{
		app:        app,
		config:     cfg,
		shutdownCh: make(chan struct{}),
		queue:      NewTranslationQueue(), // Initialize translation queue
	}

	// Suppress Fiber's default logger output before any routes
	app.Server().Logger = &DiscardLogger{}

	// Routes
	app.Get("/health", server.health)
	app.Post("/poweron", server.poweron)
	app.Post("/poweroff", server.poweroff)
	app.Get("/ready", server.ready)
	app.Post("/compute", server.compute)

	return server
}

// health handles the /health endpoint
func (s *Server) health(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// poweron handles the /poweron endpoint
func (s *Server) poweron(c fiber.Ctx) error {
	var req PoweronRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweronInvalidParams, "Invalid JSON: "+err.Error()))
	}

	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweronInvalidParams, "path is required"))
	}

	// Resolve path: if absolute, use as-is; otherwise join with work directory
	var fullPath string
	if filepath.IsAbs(req.Path) {
		fullPath = req.Path
	} else {
		fullPath = filepath.Join(s.config.WorkDir, req.Path)
	}

	// Check if path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(
			NewErrorResponse(CodePoweronPathNotExists, "path does not exist: "+fullPath))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Unload existing engine if any
	if s.translator != nil {
		if err := s.translator.Close(context.Background()); err != nil {
			Warn("Failed to close existing translator: %v", err)
		}
		s.translator = nil
	}
	if s.loadedFiles != nil {
		s.loadedFiles.Close()
		s.loadedFiles = nil
	}

	// Create translator using model directory
	ctx := context.Background()
	config := EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := CreateTranslator(ctx, config)
	if err != nil {
		// Determine error code based on error message
		errMsg := err.Error()
		if containsAny(errMsg, "not found", "missing") {
			return c.Status(fiber.StatusBadRequest).JSON(
				NewErrorResponse(CodePoweronIncompleteFiles, err.Error()))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			NewErrorResponse(CodePoweronInternalError, err.Error()))
	}

	s.translator = translator
	s.loadedFiles = loadedFiles

	// Update the queue's translator
	s.queue.SetTranslator(translator)

	return c.JSON(NewSuccessResponse(fiber.Map{"message": "Engine loaded successfully"}))
}

// poweroff handles the /poweroff endpoint
func (s *Server) poweroff(c fiber.Ctx) error {
	var req PoweroffRequest
	if err := c.Bind().JSON(&req); err != nil {
		// If no body provided, use defaults
		req.Time = 0
		req.Force = false
	}

	// Validate parameters
	if req.Time < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweroffInvalidParams, "time must be non-negative"))
	}

	s.isShuttingDown.Store(true)

	// Handle shutdown in goroutine
	go func() {
		// Wait for specified time
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		// If not force shutdown, wait for active requests
		if !req.Force {
			// Wait for active requests to complete (with timeout)
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					Warn("Shutdown timeout reached, forcing shutdown")
					goto shutdown
				case <-ticker.C:
					if atomic.LoadInt32(&s.activeReqs) == 0 {
						goto shutdown
					}
				}
			}
		}

	shutdown:
		if err := s.app.Shutdown(); err != nil {
			Error("Error during shutdown: %v", err)
		}
		close(s.shutdownCh)
	}()

	if req.Force {
		return c.JSON(NewSuccessResponse(fiber.Map{"message": "Server is shutting down"}))
	} else {
		return c.JSON(NewErrorResponse(CodePoweroffWaitingTask, "Server is shutting down, waiting for requests to complete"))
	}
}

// ready handles the /ready endpoint
func (s *Server) ready(c fiber.Ctx) error {
	isReady := s.queue.IsReady()
	return c.JSON(NewSuccessResponse(ReadyResponse{Ready: isReady}))
}

// compute handles the /compute endpoint
func (s *Server) compute(c fiber.Ctx) error {
	// Increment active request counter
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	// Check if shutting down
	if s.isShuttingDown.Load() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			NewErrorResponse(CodeComputeInternalError, "Server is shutting down"))
	}

	var req ComputeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodeComputeInvalidParams, "Invalid JSON: "+err.Error()))
	}

	if req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodeComputeInvalidParams, "text is required"))
	}

	// Check if engine is ready
	if !s.queue.IsReady() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			NewErrorResponse(CodeComputeInvalidParams, "Engine is not ready. Please call poweron first"))
	}

	translationReq := engine.TranslationRequest{
		Text: req.Text,
		Options: engine.TranslationOptions{
			HTML: req.HTML,
		},
	}

	// Use queue for sequential processing
	ctx := context.Background()
	translatedText, err := s.queue.Translate(ctx, translationReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			NewErrorResponse(CodeComputeFailure, "Translation failed: "+err.Error()))
	}

	return c.JSON(NewSuccessResponse(ComputeResponse{TranslatedText: translatedText}))
}

// ShutdownChannel returns the shutdown channel
func (s *Server) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

// Close closes the server and releases resources
func (s *Server) Close() error {
	// Close the translation queue first
	if s.queue != nil {
		s.queue.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.translator != nil {
		if err := s.translator.Close(context.Background()); err != nil {
			return err
		}
		s.translator = nil
	}

	if s.loadedFiles != nil {
		s.loadedFiles.Close()
		s.loadedFiles = nil
	}

	return nil
}

// Listen starts the HTTP server
func (s *Server) Listen(addr string) error {
	// Fiber prints startup info to stdout, suppress it
	// by temporarily redirecting stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Start listening (this triggers Fiber's startup messages)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.app.Listen(addr)
	}()

	// Wait a bit for startup messages to be written
	time.Sleep(100 * time.Millisecond)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Discard the captured output
	io.Copy(io.Discard, r)
	r.Close()

	// Wait for and return any error
	return <-errCh
}

// containsAny checks if s contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
