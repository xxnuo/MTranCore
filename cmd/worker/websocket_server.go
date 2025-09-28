package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberlog "github.com/gofiber/fiber/v3/log"
	"github.com/kiuber/gofiber3-contrib/websocket"

	engine "github.com/xxnuo/MTranCore/engine"
)

// WebSocketServer represents the WebSocket server
type WebSocketServer struct {
	app            *fiber.App
	engineManager  *EngineManager
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

	server := &WebSocketServer{
		app:           app,
		config:        cfg,
		shutdownCh:    make(chan struct{}),
		engineManager: NewEngineManager(cfg),
	}

	// Suppress Fiber's server logger output
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
		case "reboot":
			response = s.handleReboot(msg.Data)
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

	ctx := context.Background()
	result := s.engineManager.Poweron(ctx, req.Path)

	if !result.Success {
		return WSResponse{
			Type: "poweron",
			Code: int(result.ErrorCode),
			Msg:  result.ErrorMessage,
		}
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

// handleReboot handles reboot message
func (s *WebSocketServer) handleReboot(data json.RawMessage) WSResponse {
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

// handleReady handles ready message
func (s *WebSocketServer) handleReady() WSResponse {
	isReady := s.engineManager.IsReady()

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
	if !s.engineManager.IsReady() {
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
	queue := s.engineManager.GetQueue()
	translatedText, err := queue.Translate(ctx, translationReq)
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
	return s.engineManager.Close()
}

// Listen starts the WebSocket server
func (s *WebSocketServer) Listen(addr string) error {
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
