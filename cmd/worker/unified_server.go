package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

	// Shared engine manager
	engineManager  *EngineManager
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
		app:           app,
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: NewEngineManager(cfg),
	}

	// Suppress Fiber's server logger output
	app.Server().Logger = &DiscardLogger{}

	// HTTP Routes (if enabled)
	if cfg.EnableHTTP {
		app.Get("/health", server.health)
		app.Post("/poweron", server.poweron)
		app.Post("/poweroff", server.poweroff)
		app.Post("/reboot", server.reboot)
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

	ctx := context.Background()
	result := s.engineManager.PoweronWithRequest(ctx, req)

	if !result.Success {
		statusCode := fiber.StatusInternalServerError
		switch result.ErrorCode {
		case CodePoweronInvalidParams:
			statusCode = fiber.StatusBadRequest
		case CodePoweronPathNotExists:
			statusCode = fiber.StatusNotFound
		case CodePoweronIncompleteFiles:
			statusCode = fiber.StatusBadRequest
		}
		return c.Status(statusCode).JSON(
			NewErrorResponse(result.ErrorCode, result.ErrorMessage))
	}

	// Update gRPC service if exists
	if s.grpcService != nil {
		s.grpcService.engineManager = s.engineManager
	}

	message := "Engine loaded successfully"
	if result.AlreadyLoaded {
		message = "Engine already loaded"
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"message": message}))
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
			s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
		}

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

// reboot handles the /reboot endpoint
func (s *UnifiedServer) reboot(c fiber.Ctx) error {
	var req RebootRequest
	if err := c.Bind().JSON(&req); err != nil {
		// If no body provided, use defaults
		req.Time = 0
		req.Force = false
	}

	// Validate parameters
	if req.Time < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			NewErrorResponse(CodeRebootInvalidParams, "time must be non-negative"))
	}

	// Handle reboot in goroutine if time is specified
	if req.Time > 0 {
		s.engineManager.RebootAsync(req.Time, req.Force, &s.activeReqs, nil)
		return c.JSON(NewSuccessResponse(fiber.Map{"message": "Engine will reboot in " + fmt.Sprintf("%d", req.Time) + " seconds"}))
	}

	// Immediate reboot
	ctx := context.Background()
	result := s.engineManager.Reboot(ctx, req.Force, &s.activeReqs)

	if !result.Success {
		statusCode := fiber.StatusInternalServerError
		if result.ErrorCode == CodeRebootWaitingTask {
			statusCode = fiber.StatusRequestTimeout
		}
		return c.Status(statusCode).JSON(
			NewErrorResponse(result.ErrorCode, result.ErrorMessage))
	}

	message := "Engine rebooted successfully"
	if req.Force {
		message = "Engine rebooted (forced)"
	}

	return c.JSON(NewSuccessResponse(fiber.Map{"message": message}))
}

// ready handles the /ready endpoint
func (s *UnifiedServer) ready(c fiber.Ctx) error {
	isReady := s.engineManager.IsReady()
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

	if !s.engineManager.IsReady() {
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
	queue := s.engineManager.GetQueue()
	translatedText, err := queue.Translate(ctx, translationReq)
	if err != nil {
		errMsg := err.Error()
		// Check for fatal WASM errors (module closed, exit_code, etc.)
		if strings.Contains(errMsg, "module closed") || strings.Contains(errMsg, "exit_code") {
			Error("Fatal WASM error detected, triggering reboot: %v", err)
			s.engineManager.RebootAsync(0, true, &s.activeReqs, nil)
		}

		return c.Status(fiber.StatusInternalServerError).JSON(
			NewErrorResponse(CodeComputeFailure, "Translation failed: "+errMsg))
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
		case "reboot":
			response = s.handleWSReboot(msg.Data)
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

	ctx := context.Background()
	result := s.engineManager.PoweronWithRequest(ctx, req)

	if !result.Success {
		return WSResponse{
			Type: "poweron",
			Code: int(result.ErrorCode),
			Msg:  result.ErrorMessage,
		}
	}

	// Update gRPC service if exists
	if s.grpcService != nil {
		s.grpcService.engineManager = s.engineManager
	}

	message := "Engine loaded successfully"
	if result.AlreadyLoaded {
		message = "Engine already loaded"
	}

	return WSResponse{
		Type: "poweron",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: fiber.Map{"message": message},
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
			s.engineManager.WaitForIdle(&s.activeReqs, 30*time.Second)
		}

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

// handleWSReboot handles reboot message
func (s *UnifiedServer) handleWSReboot(data json.RawMessage) WSResponse {
	var req RebootRequest
	if err := json.Unmarshal(data, &req); err != nil {
		// If no body provided, use defaults
		req.Time = 0
		req.Force = false
	}

	// Validate parameters
	if req.Time < 0 {
		return WSResponse{
			Type: "reboot",
			Code: int(CodeRebootInvalidParams),
			Msg:  "time must be non-negative",
		}
	}

	// Handle reboot in goroutine if time is specified
	if req.Time > 0 {
		s.engineManager.RebootAsync(req.Time, req.Force, &s.activeReqs, nil)
		return WSResponse{
			Type: "reboot",
			Code: int(CodeSuccess),
			Msg:  "success",
			Data: fiber.Map{"message": "Engine will reboot in " + fmt.Sprintf("%d", req.Time) + " seconds"},
		}
	}

	// Immediate reboot
	ctx := context.Background()
	result := s.engineManager.Reboot(ctx, req.Force, &s.activeReqs)

	if !result.Success {
		return WSResponse{
			Type: "reboot",
			Code: int(result.ErrorCode),
			Msg:  result.ErrorMessage,
		}
	}

	message := "Engine rebooted successfully"
	if req.Force {
		message = "Engine rebooted (forced)"
	}

	return WSResponse{
		Type: "reboot",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: fiber.Map{"message": message},
	}
}

// handleWSReady handles ready message
func (s *UnifiedServer) handleWSReady() WSResponse {
	isReady := s.engineManager.IsReady()
	return WSResponse{
		Type: "ready",
		Code: int(CodeSuccess),
		Msg:  "success",
		Data: ReadyResponse{Ready: isReady},
	}
}

// handleWSCompute handles compute message
func (s *UnifiedServer) handleWSCompute(data json.RawMessage) WSResponse {
	Debug("[DEBUG-WS] handleWSCompute: starting")
	atomic.AddInt32(&s.activeReqs, 1)
	defer atomic.AddInt32(&s.activeReqs, -1)

	if s.isShuttingDown.Load() {
		Debug("[DEBUG-WS] handleWSCompute: server is shutting down")
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInternalError),
			Msg:  "Server is shutting down",
		}
	}

	var req ComputeRequest
	if err := json.Unmarshal(data, &req); err != nil {
		Debug("[DEBUG-WS] handleWSCompute: invalid JSON: %v", err)
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "Invalid JSON: " + err.Error(),
		}
	}
	Debug("[DEBUG-WS] handleWSCompute: req.Text length=%d, req.HTML=%v", len(req.Text), req.HTML)

	if req.Text == "" {
		Debug("[DEBUG-WS] handleWSCompute: text is empty")
		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeInvalidParams),
			Msg:  "text is required",
		}
	}

	isReady := s.engineManager.IsReady()
	Debug("[DEBUG-WS] handleWSCompute: engineManager.IsReady()=%v", isReady)
	if !isReady {
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
	Debug("[DEBUG-WS] handleWSCompute: getting queue")
	queue := s.engineManager.GetQueue()
	Debug("[DEBUG-WS] handleWSCompute: queue=%v, calling Translate", queue)
	translatedText, err := queue.Translate(ctx, translationReq)
	Debug("[DEBUG-WS] handleWSCompute: Translate returned, err=%v", err)
	if err != nil {
		errMsg := err.Error()
		// Check for fatal WASM errors (module closed, exit_code, etc.)
		if strings.Contains(errMsg, "module closed") || strings.Contains(errMsg, "exit_code") {
			Error("Fatal WASM error detected (WebSocket), triggering reboot: %v", err)
			s.engineManager.RebootAsync(0, true, &s.activeReqs, nil)
		}

		return WSResponse{
			Type: "compute",
			Code: int(CodeComputeFailure),
			Msg:  "Translation failed: " + errMsg,
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
	grpc.engineManager = s.engineManager
}

// ShutdownChannel returns the shutdown channel
func (s *UnifiedServer) ShutdownChannel() <-chan struct{} {
	return s.shutdownCh
}

// Close closes the server and releases resources
func (s *UnifiedServer) Close() error {
	return s.engineManager.Close()
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
