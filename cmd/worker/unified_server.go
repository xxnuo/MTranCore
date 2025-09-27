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
	fiberlog "github.com/gofiber/fiber/v3/log"
	"github.com/kiuber/gofiber3-contrib/websocket"

	engine "github.com/xxnuo/MTranCore/engine"
)

// UnifiedServer represents a unified server handling HTTP, WebSocket, and gRPC
type UnifiedServer struct {
	// Fiber app for HTTP and WebSocket
	app *fiber.App

	// gRPC service
	grpcService *GRPCServer

	// Shared translator and queue
	translator     *engine.Translator
	loadedFiles    *engine.LoadedFiles
	queue          *TranslationQueue
	mu             sync.RWMutex
	shutdownCh     chan struct{}
	config         *Config
	activeReqs     int32 // atomic counter for active requests
	isShuttingDown atomic.Bool
}

// NewUnifiedServer creates a new unified server instance
func NewUnifiedServer(cfg *Config) *UnifiedServer {
	// Redirect Fiber's log output to our standard logger
	if globalLogger != nil {
		fiberLogger := globalLogger.GetWriter(LogLevelInfo)
		fiberlog.SetOutput(fiberLogger)
	} else {
		// Discard Fiber logs if no logger is set
		fiberlog.SetOutput(io.Discard)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(NewErrorResponse(CodePoweronUnknownError, err.Error()))
		},
	})

	server := &UnifiedServer{
		app:        app,
		config:     cfg,
		shutdownCh: make(chan struct{}),
		queue:      NewTranslationQueue(),
	}

	// Suppress Fiber's server logger output
	app.Server().Logger = &DiscardLogger{}

	// HTTP Routes (if enabled)
	if cfg.EnableHTTP {
		app.Get("/health", server.health)
		app.Post("/poweron", server.poweron)
		app.Post("/poweroff", server.poweroff)
		app.Get("/ready", server.ready)
		app.Post("/compute", server.compute)
	}

	// WebSocket Routes (if enabled)
	if cfg.EnableWebSocket {
		app.Use("/ws", func(c fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})
		app.Get("/ws", websocket.New(server.handleWebSocket))
	}

	return server
}

// GetApp returns the fiber app instance
func (s *UnifiedServer) GetApp() *fiber.App {
	return s.app
}

// ===== HTTP Handlers =====

// health handles the /health endpoint
func (s *UnifiedServer) health(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}

// poweron handles the /poweron endpoint
func (s *UnifiedServer) poweron(c fiber.Ctx) error {
	var req PoweronRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweronInvalidParams, "Invalid JSON: "+err.Error()))
	}

	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweronInvalidParams, "path is required"))
	}

	// Resolve path
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

	// Check if engine is already loaded
	if s.translator != nil {
		// Engine already loaded, return success immediately
		return c.JSON(NewSuccessResponse(fiber.Map{"message": "Engine already loaded"}))
	}

	// Create translator
	ctx := context.Background()
	config := engine.EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
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
	s.queue.SetTranslator(translator)

	// Update gRPC service translator if exists
	if s.grpcService != nil {
		s.grpcService.mu.Lock()
		if s.grpcService.translator != nil {
			s.grpcService.translator.Close(context.Background())
		}
		if s.grpcService.loadedFiles != nil {
			s.grpcService.loadedFiles.Close()
		}
		s.grpcService.translator = translator
		s.grpcService.loadedFiles = loadedFiles
		s.grpcService.mu.Unlock()
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"message": "Engine loaded successfully"}))
}

// poweroff handles the /poweroff endpoint
func (s *UnifiedServer) poweroff(c fiber.Ctx) error {
	var req PoweroffRequest
	if err := c.Bind().JSON(&req); err != nil {
		req.Time = 0
		req.Force = false
	}

	if req.Time < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodePoweroffInvalidParams, "time must be non-negative"))
	}

	s.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
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
	}
	return c.JSON(NewErrorResponse(CodePoweroffWaitingTask, "Server is shutting down, waiting for requests to complete"))
}

// ready handles the /ready endpoint
func (s *UnifiedServer) ready(c fiber.Ctx) error {
	isReady := s.queue.IsReady()
	return c.JSON(NewSuccessResponse(ReadyResponse{Ready: isReady}))
}

// compute handles the /compute endpoint
func (s *UnifiedServer) compute(c fiber.Ctx) error {
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

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

	ctx := context.Background()
	translatedText, err := s.queue.Translate(ctx, translationReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			NewErrorResponse(CodeComputeFailure, "Translation failed: "+err.Error()))
	}

	// Return plain text on success
	c.Set("Content-Type", "text/plain; charset=utf-8")
	return c.SendString(translatedText)
}

// ===== WebSocket Handlers =====

// handleWebSocket handles websocket connections
func (s *UnifiedServer) handleWebSocket(c *websocket.Conn) {
	defer c.Close()

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

		var response WSResponse
		switch msg.Type {
		case "poweron":
			response = s.handleWSPoweron(msg.Data)
		case "poweroff":
			response = s.handleWSPoweroff(msg.Data)
		case "ready":
			response = s.handleWSReady()
		case "compute":
			response = s.handleWSCompute(msg.Data)
		default:
			response = WSResponse{
				Type: msg.Type,
				Code: int(CodePoweronUnknownError),
				Msg:  "Unknown message type: " + msg.Type,
			}
		}

		if err := c.WriteJSON(response); err != nil {
			Error("[WebSocket] Error sending response: %v", err)
			break
		}

		if msg.Type == "poweroff" {
			break
		}
	}

	Debug("[WebSocket] Connection closed from %s", c.RemoteAddr().String())
}

// handleWSPoweron handles poweron message
func (s *UnifiedServer) handleWSPoweron(data json.RawMessage) WSResponse {
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

	var fullPath string
	if filepath.IsAbs(req.Path) {
		fullPath = req.Path
	} else {
		fullPath = filepath.Join(s.config.WorkDir, req.Path)
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return WSResponse{
			Type: "poweron",
			Code: int(CodePoweronPathNotExists),
			Msg:  "path does not exist: " + fullPath,
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if engine is already loaded
	if s.translator != nil {
		// Engine already loaded, return success immediately
		return WSResponse{
			Type: "poweron",
			Code: int(CodeSuccess),
			Msg:  "success",
			Data: fiber.Map{"message": "Engine already loaded"},
		}
	}

	ctx := context.Background()
	config := engine.EngineConfig{ModelDir: fullPath}
	translator, loadedFiles, err := engine.CreateTranslator(ctx, config)
	if err != nil {
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
	s.queue.SetTranslator(translator)

	// Update gRPC service translator if exists
	if s.grpcService != nil {
		s.grpcService.mu.Lock()
		if s.grpcService.translator != nil {
			s.grpcService.translator.Close(context.Background())
		}
		if s.grpcService.loadedFiles != nil {
			s.grpcService.loadedFiles.Close()
		}
		s.grpcService.translator = translator
		s.grpcService.loadedFiles = loadedFiles
		s.grpcService.mu.Unlock()
	}

	return WSResponse{
		Type: "poweron",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: fiber.Map{"message": "Engine loaded successfully"},
	}
}

// handleWSPoweroff handles poweroff message
func (s *UnifiedServer) handleWSPoweroff(data json.RawMessage) WSResponse {
	var req PoweroffRequest
	if err := json.Unmarshal(data, &req); err != nil {
		req.Time = 0
		req.Force = false
	}

	if req.Time < 0 {
		return WSResponse{
			Type: "poweroff",
			Code: int(CodePoweroffInvalidParams),
			Msg:  "time must be non-negative",
		}
	}

	s.isShuttingDown.Store(true)

	go func() {
		if req.Time > 0 {
			time.Sleep(time.Duration(req.Time) * time.Second)
		}

		if !req.Force {
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

// handleWSReady handles ready message
func (s *UnifiedServer) handleWSReady() WSResponse {
	isReady := s.queue.IsReady()
	return WSResponse{
		Type: "ready",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: ReadyResponse{Ready: isReady},
	}
}

// handleWSCompute handles compute message
func (s *UnifiedServer) handleWSCompute(data json.RawMessage) WSResponse {
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

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

// SetGRPCService sets the gRPC service reference for shared state
func (s *UnifiedServer) SetGRPCService(grpc *GRPCServer) {
	s.grpcService = grpc
}

// ShutdownChannel returns the shutdown channel
func (s *UnifiedServer) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

// Close closes the server and releases resources
func (s *UnifiedServer) Close() error {
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

// Listen starts the unified server (HTTP/WebSocket only, gRPC handled separately)
func (s *UnifiedServer) Listen(addr string) error {
	// Temporarily redirect stdout to suppress Fiber's startup banner
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Start listening in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.app.Listen(addr, fiber.ListenConfig{
			DisableStartupMessage: true,
		})
	}()

	// Wait a bit for startup messages to be written
	time.Sleep(100 * time.Millisecond)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Discard the captured output
	io.Copy(io.Discard, r)
	r.Close()

	// Return any error
	return <-errCh
}
