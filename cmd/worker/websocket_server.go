package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/kiuber/gofiber3-contrib/websocket"

	engine "github.com/xxnuo/MTranCore/engine"
)

// WebSocketServer represents the WebSocket server
type WebSocketServer struct {
	app            *fiber.App
	translator     *engine.Translator
	loadedFiles    *engine.LoadedFiles
	queue          *TranslationQueue // Request queue for sequential processing
	mu             sync.RWMutex
	shutdownCh     chan struct{}
	config         *Config
	activeReqs     int32 // atomic counter for active requests
	isShuttingDown atomic.Bool
}

// WSMessage represents a websocket message
type WSMessage struct {
	Type string          `json:"type"` // "poweron", "poweroff", "ready", "compute"
	Data json.RawMessage `json:"data"`
}

// WSResponse represents a websocket response
type WSResponse struct {
	Type string      `json:"type"` // same as request type
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// NewWebSocketServer creates a new WebSocket server instance
func NewWebSocketServer(cfg *Config) *WebSocketServer {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(NewErrorResponse(CodePoweronUnknownError, err.Error()))
		},
	})

	server := &WebSocketServer{
		app:        app,
		config:     cfg,
		shutdownCh: make(chan struct{}),
		queue:      NewTranslationQueue(), // Initialize translation queue
	}

	// Suppress Fiber's default logger output before any routes
	app.Server().Logger = &DiscardLogger{}

	// WebSocket upgrade middleware
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket handler
	app.Get("/ws", websocket.New(server.handleWebSocket))

	// Health check endpoint
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	return server
}

// handleWebSocket handles websocket connections
func (s *WebSocketServer) handleWebSocket(c *websocket.Conn) {
	defer c.Close()

	// Connection established
	Debug("[WebSocket] New connection from %s", c.RemoteAddr().String())

	for {
		var msg WSMessage
		err := c.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Error("[WebSocket] Error reading message: %v", err)
			}
			break
		}

		// Handle message based on type
		var response WSResponse
		switch msg.Type {
		case "poweron":
			response = s.handlePoweron(msg.Data)
		case "poweroff":
			response = s.handlePoweroff(msg.Data)
		case "ready":
			response = s.handleReady()
		case "compute":
			response = s.handleCompute(msg.Data)
		default:
			response = WSResponse{
				Type: msg.Type,
				Code: int(CodePoweronUnknownError),
				Msg:  "Unknown message type: " + msg.Type,
			}
		}

		// Send response
		if err := c.WriteJSON(response); err != nil {
			Error("[WebSocket] Error sending response: %v", err)
			break
		}

		// If poweroff, close connection
		if msg.Type == "poweroff" {
			break
		}
	}

	Debug("[WebSocket] Connection closed from %s", c.RemoteAddr().String())
}

// handlePoweron handles poweron message
func (s *WebSocketServer) handlePoweron(data json.RawMessage) WSResponse {
	var req PoweronRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return WSResponse{
			Type: "poweron",
			Code: int(CodePoweronInvalidParams),
			Msg:  "Invalid JSON: " + err.Error(),
		}
	}

	if req.Path == "" {
		return WSResponse{
			Type: "poweron",
			Code: int(CodePoweronInvalidParams),
			Msg:  "path is required",
		}
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
		return WSResponse{
			Type: "poweron",
			Code: int(CodePoweronPathNotExists),
			Msg:  "path does not exist: " + fullPath,
		}
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
	config := engine.EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
		// Determine error code based on error message
		errMsg := err.Error()
		if containsAny(errMsg, "not found", "missing") {
			return WSResponse{
				Type: "poweron",
				Code: int(CodePoweronIncompleteFiles),
				Msg:  err.Error(),
			}
		}
		return WSResponse{
			Type: "poweron",
			Code: int(CodePoweronInternalError),
			Msg:  err.Error(),
		}
	}

	s.translator = translator
	s.loadedFiles = loadedFiles

	// Update the queue's translator
	s.queue.SetTranslator(translator)

	return WSResponse{
		Type: "poweron",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: fiber.Map{"message": "Engine loaded successfully"},
	}
}

// handlePoweroff handles poweroff message
func (s *WebSocketServer) handlePoweroff(data json.RawMessage) WSResponse {
	var req PoweroffRequest
	if err := json.Unmarshal(data, &req); err != nil {
		// If no body provided, use defaults
		req.Time = 0
		req.Force = false
	}

	// Validate parameters
	if req.Time < 0 {
		return WSResponse{
			Type: "poweroff",
			Code: int(CodePoweroffInvalidParams),
			Msg:  "time must be non-negative",
		}
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
		return WSResponse{
			Type: "poweroff",
			Code: int(CodeSuccess),
			Msg:  "success",
			Data: fiber.Map{"message": "Server is shutting down"},
		}
	}
	return WSResponse{
		Type: "poweroff",
		Code: int(CodePoweroffWaitingTask),
		Msg:  "Server is shutting down, waiting for requests to complete",
	}
}

// handleReady handles ready message
func (s *WebSocketServer) handleReady() WSResponse {
	isReady := s.queue.IsReady()

	return WSResponse{
		Type: "ready",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: ReadyResponse{Ready: isReady},
	}
}

// handleCompute handles compute message
func (s *WebSocketServer) handleCompute(data json.RawMessage) WSResponse {
	// Increment active request counter
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	// Check if shutting down
	if s.isShuttingDown.Load() {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInternalError),
			Msg:  "Server is shutting down",
		}
	}

	var req ComputeRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "Invalid JSON: " + err.Error(),
		}
	}

	if req.Text == "" {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "text is required",
		}
	}

	// Check if engine is ready
	if !s.queue.IsReady() {
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "Engine is not ready. Please call poweron first",
		}
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
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeFailure),
			Msg:  "Translation failed: " + err.Error(),
		}
	}

	return WSResponse{
		Type: "compute",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: ComputeResponse{TranslatedText: translatedText},
	}
}

// ShutdownChannel returns the shutdown channel
func (s *WebSocketServer) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

// Close closes the server and releases resources
func (s *WebSocketServer) Close() error {
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

// Listen starts the WebSocket server
func (s *WebSocketServer) Listen(addr string) error {
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
